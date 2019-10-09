package test

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/require"
)

func TestLeadDashboard(t *testing.T) {
	//t.Parallel()
	// Create a namespace for the test
	expectedNamespace := "terratest-test-namespace"
	expectedTillerServiceAccountName := "tiller"

	terraformNamespaceOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../modules/common/namespace",

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"namespace": expectedNamespace,
		},
	}

	// Start Dashboard test
	expectedRootZoneName := "test-rootzonename"
	expectedCluster := "test-cluster"
	expectedDashboardVersion := "0.2.2"

	terraformDashboardOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../modules/lead/dashboard/",

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"root_zone_name":    expectedRootZoneName,
			"cluster":           expectedCluster,
			"namespace":         expectedNamespace,
			"dashboard_version": expectedDashboardVersion,
		},

		// Variables to pass to our Terraform code using -var-file options
		// VarFiles: []string{"varfile.tfvars"},

		// Disable colors in Terraform commands so its easier to parse stdout/stderr
		NoColor: true,
	}

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	defer terraform.Destroy(t, terraformDashboardOptions)
	defer terraform.Destroy(t, terraformNamespaceOptions)
	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	terraform.InitAndApply(t, terraformNamespaceOptions)
	terraform.InitAndApply(t, terraformDashboardOptions)

	// Setup the kubectl config and context. Here we choose to create a new one because we will be manipulating the
	// entries to be able to add a new authentication option.
	tmpConfigPath := k8s.CopyHomeKubeConfigToTemp(t)
	defer os.Remove(tmpConfigPath)
	options := k8s.NewKubectlOptions("", tmpConfigPath)

	// Namespace auth to access dashboard service
	options.Namespace = expectedNamespace
	token := k8s.GetServiceAccountAuthToken(t, options, expectedTillerServiceAccountName)

	require.NoError(t, k8s.AddConfigContextForServiceAccountE(
		t,
		options,
		expectedTillerServiceAccountName, // for this test we will name the context after the ServiceAccount
		expectedTillerServiceAccountName,
		token,
	))
	// Section below is used when making calls to k8s.
	// serviceAccountKubectlOptions := k8s.NewKubectlOptions(expectedTillerServiceAccountName, tmpConfigPath)

	// Next we wait until the service is available. This will wait up to 10 seconds for the service to become available,
	// to ensure that we can access it.
	// k8s.WaitUntilServiceAvailable(t, serviceAccountKubectlOptions, serviceName, 10, 1*time.Second)

}
