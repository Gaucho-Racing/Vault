package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/gaucho-racing/ulid-go"
	"github.com/gaucho-racing/vault/vault/database"
	"github.com/gaucho-racing/vault/vault/model"
	"github.com/gaucho-racing/vault/vault/pkg/kubernetes"
	"gorm.io/gorm"
)

var ErrKubernetesClusterNameRequired = errors.New("kubernetes cluster name is required")
var ErrKubernetesClusterNameInvalid = errors.New("kubernetes cluster name must contain only lowercase letters, numbers, hyphens, and underscores")
var ErrKubernetesClusterIssuerRequired = errors.New("kubernetes cluster issuer is required")
var ErrKubernetesClusterIssuerInvalid = errors.New("kubernetes cluster issuer must be an http or https URL")
var ErrKubernetesClusterAudienceInvalid = errors.New("kubernetes cluster audience cannot contain whitespace")
var ErrKubernetesClusterNotTrusted = errors.New("kubernetes cluster is not trusted")
var ErrKubernetesClusterInUse = errors.New("kubernetes cluster is used by an access rule")
var ErrKubernetesClusterVerificationFailed = errors.New("kubernetes cluster verification failed")

var kubernetesAudiencePattern = regexp.MustCompile(`^\S+$`)

type KubernetesVerifiedIdentity struct {
	Cluster model.KubernetesCluster
	Claims  kubernetes.Claims
}

var kubernetesVerifierStore = struct {
	sync.Mutex
	items map[string]*kubernetes.Verifier
}{items: map[string]*kubernetes.Verifier{}}

func GetAllKubernetesClusters() ([]model.KubernetesCluster, error) {
	clusters := []model.KubernetesCluster{}
	if err := database.DB.Order("name ASC").Find(&clusters).Error; err != nil {
		return []model.KubernetesCluster{}, err
	}
	return clusters, nil
}

func GetKubernetesClusterByID(id string) (model.KubernetesCluster, error) {
	var cluster model.KubernetesCluster
	if err := database.DB.Where("id = ?", id).First(&cluster).Error; err != nil {
		return model.KubernetesCluster{}, err
	}
	return cluster, nil
}

func CreateKubernetesCluster(ctx context.Context, cluster model.KubernetesCluster) (model.KubernetesCluster, error) {
	if cluster.ID == "" {
		cluster.ID = ulid.Make().Prefixed("k8scluster")
	}
	if err := VerifyKubernetesClusterConfig(ctx, &cluster); err != nil {
		return model.KubernetesCluster{}, err
	}
	if err := database.DB.Create(&cluster).Error; err != nil {
		return model.KubernetesCluster{}, err
	}
	return cluster, nil
}

func UpdateKubernetesCluster(ctx context.Context, cluster model.KubernetesCluster) (model.KubernetesCluster, error) {
	if err := VerifyKubernetesClusterConfig(ctx, &cluster); err != nil {
		return model.KubernetesCluster{}, err
	}
	if err := database.DB.Save(&cluster).Error; err != nil {
		return model.KubernetesCluster{}, err
	}
	return cluster, nil
}

func DeleteKubernetesCluster(id string) error {
	rules := []model.KubernetesSecretRule{}
	if err := database.DB.Find(&rules).Error; err != nil {
		return err
	}
	for _, rule := range rules {
		if stringSliceContains(rule.ClusterIDs, id) {
			return ErrKubernetesClusterInUse
		}
	}

	result := database.DB.Where("id = ?", id).Delete(&model.KubernetesCluster{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func VerifyKubernetesToken(ctx context.Context, token string) (KubernetesVerifiedIdentity, error) {
	issuer, err := kubernetes.UnverifiedIssuer(token)
	if err != nil {
		return KubernetesVerifiedIdentity{}, err
	}

	cluster, err := getEnabledKubernetesClusterByIssuer(issuer)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return KubernetesVerifiedIdentity{}, ErrKubernetesClusterNotTrusted
		}
		return KubernetesVerifiedIdentity{}, err
	}

	claims, err := kubernetesVerifierForCluster(cluster).Verify(ctx, token)
	if err != nil {
		return KubernetesVerifiedIdentity{}, err
	}
	return KubernetesVerifiedIdentity{Cluster: cluster, Claims: claims}, nil
}

func VerifyKubernetesClusterConfig(ctx context.Context, cluster *model.KubernetesCluster) error {
	if err := normalizeKubernetesCluster(cluster); err != nil {
		return err
	}
	if err := kubernetes.NewVerifier(cluster.Issuer, cluster.Audience, nil).Check(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrKubernetesClusterVerificationFailed, err)
	}
	return nil
}

func normalizeKubernetesCluster(cluster *model.KubernetesCluster) error {
	cluster.Name = strings.ToLower(strings.TrimSpace(cluster.Name))
	cluster.Issuer = strings.TrimRight(strings.TrimSpace(cluster.Issuer), "/")
	cluster.Audience = strings.TrimSpace(cluster.Audience)
	if cluster.Audience == "" {
		cluster.Audience = kubernetes.DefaultAudience
	}

	if cluster.Name == "" {
		return ErrKubernetesClusterNameRequired
	}
	if !appSecretIdentifierPattern.MatchString(cluster.Name) {
		return ErrKubernetesClusterNameInvalid
	}
	if cluster.Issuer == "" {
		return ErrKubernetesClusterIssuerRequired
	}
	issuerURL, err := url.Parse(cluster.Issuer)
	if err != nil || issuerURL.Host == "" || (issuerURL.Scheme != "https" && issuerURL.Scheme != "http") {
		return ErrKubernetesClusterIssuerInvalid
	}
	if !kubernetesAudiencePattern.MatchString(cluster.Audience) {
		return ErrKubernetesClusterAudienceInvalid
	}
	return nil
}

func getEnabledKubernetesClusterByIssuer(issuer string) (model.KubernetesCluster, error) {
	var cluster model.KubernetesCluster
	if err := database.DB.
		Where("enabled = ? AND issuer = ?", true, strings.TrimRight(strings.TrimSpace(issuer), "/")).
		First(&cluster).Error; err != nil {
		return model.KubernetesCluster{}, err
	}
	return cluster, nil
}

func kubernetesVerifierForCluster(cluster model.KubernetesCluster) *kubernetes.Verifier {
	key := strings.Join([]string{cluster.ID, cluster.Issuer, cluster.Audience}, "\x00")

	kubernetesVerifierStore.Lock()
	defer kubernetesVerifierStore.Unlock()
	verifier := kubernetesVerifierStore.items[key]
	if verifier == nil {
		verifier = kubernetes.NewVerifier(cluster.Issuer, cluster.Audience, nil)
		kubernetesVerifierStore.items[key] = verifier
	}
	return verifier
}

func stringSliceContains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
