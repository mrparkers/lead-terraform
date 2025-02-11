provider "helm" {
  alias = "system"
}

provider "helm" {
  alias = "toolchain"
}

data "helm_repository" "liatrio" {
  name = "lead.prod.liatr.io"
  url  = "https://artifactory.toolchain.lead.prod.liatr.io/artifactory/helm/"
  provider   = helm.system
}

resource "helm_release" "operator_toolchain_definition" {
  count      = var.enable_operators ? 1 : 0
  repository = data.helm_repository.liatrio.metadata[0].name
  name       = "operator-toolchain-definition"
  chart      = "operator-toolchain-definition"
  version    = var.sdm_version
  namespace  = var.system_namespace
  provider   = helm.system
}

data "template_file" "operator_toolchain_values" {
  template = file("${path.module}/operator-toolchain-values.tpl")

  vars = {
    image_tag           = "v${var.sdm_version}"
    cluster             = var.cluster
    namespace           = var.namespace
    cluster_domain      = "${var.cluster}.${var.root_zone_name}"
    product_version     = var.product_version
    workspace_role      = var.workspace_role_name
    region              = var.region
    product_stack       = var.product_stack
    product_vars        = jsonencode(var.product_vars)

    slack_service_account_annotations   = jsonencode(var.operator_slack_service_account_annotations)
    jenkins_service_account_annotations = jsonencode(var.operator_jenkins_service_account_annotations)
  }
}

resource "helm_release" "operator_toolchain" {
  count      = var.enable_operators ? 1 : 0
  repository = data.helm_repository.liatrio.metadata[0].name
  name       = "operator-toolchain"
  chart      = "operator-toolchain"
  version    = var.sdm_version
  namespace  = var.namespace
  provider   = helm.toolchain
  depends_on = [helm_release.operator_toolchain_definition]

  values = [data.template_file.operator_toolchain_values.rendered]
}

resource "kubernetes_secret" "operator_slack_config" {
  metadata {
    name      = "operator-slack-config"
    namespace = var.namespace

    labels = {
      "app.kubernetes.io/name"       = "operator-slack"
      "app.kubernetes.io/instance"   = "operator-slack"
      "app.kubernetes.io/component"  = "operator-slack"
      "app.kubernetes.io/managed-by" = "Terraform"
    }

    annotations = {
      "source-repo" = "https://github.com/liatrio/lead-toolchain"
    }
  }

  type = "Opaque"

  data = {
    "slack_config" = <<EOF
clientSigningSecret=${var.slack_client_signing_secret}
botToken=${var.slack_bot_token}
EOF

  }
}
