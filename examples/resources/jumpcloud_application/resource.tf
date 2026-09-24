# SAML application
resource "jumpcloud_application" "finance" {
  name          = "finance-portal"
  display_label = "Finance Portal"
  sso_url       = "https://finance.example.com/sso"

  logo = {
    url = "https://finance.example.com/logo.png"
  }

  sso = {
    type         = "saml"
    sp_entity_id = "https://finance.example.com/saml/metadata"
    acs_url      = "https://finance.example.com/saml/acs"
  }
}

# OIDC application with SCIM provisioning
resource "jumpcloud_application" "wiki" {
  name          = "wiki"
  display_label = "Company Wiki"

  sso = {
    type          = "oidc"
    redirect_uris = ["https://wiki.example.com/oauth/callback"]
  }

  provision = {
    enabled  = true
    type     = "scim"
    base_url = "https://wiki.example.com/scim/v2"
    token    = var.wiki_scim_token
  }
}
