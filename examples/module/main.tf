# linux function app
resource "azurerm_linux_function_app" "func" {
  for_each = var.instance.type == "linux" ? {
    (var.instance.name) = var.instance
  } : {}

  resource_group_name = coalesce(
    lookup(
      var.instance, "resource_group", null
    ), var.resource_group
  )

  location = coalesce(
    lookup(var.instance, "location", null
    ), var.location
  )

  name                                           = var.instance.name
  service_plan_id                                = var.instance.service_plan_id
  storage_account_name                           = var.instance.storage_account_name
  storage_account_access_key                     = var.instance.storage_account_access_key
  https_only                                     = var.instance.https_only
  zip_deploy_file                                = var.instance.zip_deploy_file
  enabled                                        = var.instance.enabled
  builtin_logging_enabled                        = var.instance.builtin_logging_enabled
  client_certificate_mode                        = var.instance.client_certificate_mode
  daily_memory_time_quota                        = var.instance.daily_memory_time_quota
  virtual_network_subnet_id                      = var.instance.virtual_network_subnet_id
  client_certificate_enabled                     = var.instance.client_certificate_enabled
  functions_extension_version                    = var.instance.functions_extension_version
  storage_key_vault_secret_id                    = var.instance.storage_key_vault_secret_id
  content_share_force_disabled                   = var.instance.content_share_force_disabled
  public_network_access_enabled                  = var.instance.public_network_access_enabled
  storage_uses_managed_identity                  = var.instance.storage_uses_managed_identity
  vnet_image_pull_enabled                        = var.instance.vnet_image_pull_enabled
  key_vault_reference_identity_id                = var.instance.key_vault_reference_identity_id
  client_certificate_exclusion_paths             = var.instance.client_certificate_exclusion_paths
  ftp_publish_basic_authentication_enabled       = var.instance.ftp_publish_basic_authentication_enabled
  webdeploy_publish_basic_authentication_enabled = var.instance.webdeploy_publish_basic_authentication_enabled
  virtual_network_backup_restore_enabled         = var.instance.virtual_network_backup_restore_enabled
  app_settings                                   = var.instance.app_settings

  tags = try(
    var.instance.tags, var.tags, null
  )

  dynamic "identity" {
    for_each = lookup(var.instance, "identity", null) != null ? [var.instance.identity] : []

    content {
      type         = identity.value.type
      identity_ids = identity.value.identity_ids
    }
  }

  dynamic "auth_settings_v2" {
    for_each = lookup(each.value, "auth_settings_v2", null) != null ? [lookup(each.value, "auth_settings_v2")] : []

    content {
      auth_enabled                            = auth_settings_v2.value.auth_enabled
      runtime_version                         = auth_settings_v2.value.runtime_version
      config_file_path                        = auth_settings_v2.value.config_file_path
      require_authentication                  = auth_settings_v2.value.require_authentication
      unauthenticated_action                  = auth_settings_v2.value.unauthenticated_action
      default_provider                        = auth_settings_v2.value.default_provider
      excluded_paths                          = auth_settings_v2.value.excluded_paths
      require_https                           = auth_settings_v2.value.require_https
      http_route_api_prefix                   = auth_settings_v2.value.http_route_api_prefix
      forward_proxy_convention                = auth_settings_v2.value.forward_proxy_convention
      forward_proxy_custom_host_header_name   = auth_settings_v2.value.forward_proxy_custom_host_header_name
      forward_proxy_custom_scheme_header_name = auth_settings_v2.value.forward_proxy_custom_scheme_header_name

      login {
        token_store_enabled               = auth_settings_v2.value.login.token_store_enabled
        token_refresh_extension_time      = auth_settings_v2.value.login.token_refresh_extension_time
        token_store_path                  = auth_settings_v2.value.login.token_store_path
        token_store_sas_setting_name      = auth_settings_v2.value.login.token_store_sas_setting_name
        preserve_url_fragments_for_logins = auth_settings_v2.value.login.preserve_url_fragments_for_logins
        allowed_external_redirect_urls    = auth_settings_v2.value.login.allowed_external_redirect_urls
        cookie_expiration_convention      = auth_settings_v2.value.login.cookie_expiration_convention
        cookie_expiration_time            = auth_settings_v2.value.login.cookie_expiration_time
        validate_nonce                    = auth_settings_v2.value.login.validate_nonce
        nonce_expiration_time             = auth_settings_v2.value.login.nonce_expiration_time
        logout_endpoint                   = auth_settings_v2.value.login.logout_endpoint
      }

      dynamic "apple_v2" {
        for_each = lookup(auth_settings_v2.value, "apple_v2", null) != null ? [lookup(auth_settings_v2.value, "apple_v2")] : []

        content {
          client_id                  = apple_v2.value.client_id
          client_secret_setting_name = apple_v2.value.client_secret_setting_name
          login_scopes               = apple_v2.value.login_scopes
        }
      }

      dynamic "active_directory_v2" {
        for_each = lookup(auth_settings_v2.value, "active_directory_v2", null) != null ? [lookup(auth_settings_v2.value, "active_directory_v2")] : []

        content {
          client_id                            = active_directory_v2.value.client_id
          tenant_auth_endpoint                 = active_directory_v2.value.tenant_auth_endpoint
          client_secret_setting_name           = active_directory_v2.value.client_secret_setting_name
          client_secret_certificate_thumbprint = active_directory_v2.value.client_secret_certificate_thumbprint
          jwt_allowed_groups                   = active_directory_v2.value.jwt_allowed_groups
          jwt_allowed_client_applications      = active_directory_v2.value.jwt_allowed_client_applications
          www_authentication_disabled          = active_directory_v2.value.www_authentication_disabled
          allowed_audiences                    = active_directory_v2.value.allowed_audiences
          allowed_groups                       = active_directory_v2.value.allowed_groups
          allowed_identities                   = active_directory_v2.value.allowed_identities
          login_parameters                     = active_directory_v2.value.login_parameters
          allowed_applications                 = active_directory_v2.value.allowed_applications
        }
      }

      dynamic "azure_static_web_app_v2" {
        for_each = lookup(auth_settings_v2.value, "azure_static_web_app_v2", null) != null ? [lookup(auth_settings_v2.value, "azure_static_web_app_v2")] : []

        content {
          client_id = azure_static_web_app_v2.value.client_id
        }
      }

      dynamic "custom_oidc_v2" {
        for_each = try(auth_settings_v2.value.custom_oidc_v2, {})

        content {
          name                          = custom_oidc_v2.value.name
          client_id                     = custom_oidc_v2.value.client_id
          openid_configuration_endpoint = custom_oidc_v2.value.openid_configuration_endpoint
          name_claim_type               = custom_oidc_v2.value.name_claim_type
          scopes                        = custom_oidc_v2.value.scopes
          client_credential_method      = custom_oidc_v2.value.client_credential_method
          client_secret_setting_name    = custom_oidc_v2.value.client_secret_setting_name
          authorisation_endpoint        = custom_oidc_v2.value.authorisation_endpoint
          token_endpoint                = custom_oidc_v2.value.token_endpoint
          issuer_endpoint               = custom_oidc_v2.value.issuer_endpoint
          certification_uri             = custom_oidc_v2.value.certification_uri
        }
      }

      dynamic "facebook_v2" {
        for_each = lookup(auth_settings_v2.value, "facebook_v2", null) != null ? [lookup(auth_settings_v2.value, "facebook_v2")] : []

        content {
          app_id                  = facebook_v2.value.app_id
          app_secret_setting_name = facebook_v2.value.app_secret_setting_name
          graph_api_version       = facebook_v2.value.graph_api_version
          login_scopes            = facebook_v2.value.login_scopes
        }
      }

      dynamic "github_v2" {
        for_each = lookup(auth_settings_v2.value, "github_v2", null) != null ? [lookup(auth_settings_v2.value, "github_v2")] : []

        content {
          client_id                  = github_v2.value.client_id
          client_secret_setting_name = github_v2.value.client_secret_setting_name
          login_scopes               = github_v2.value.login_scopes
        }
      }

      dynamic "google_v2" {
        for_each = lookup(auth_settings_v2.value, "google_v2", null) != null ? [lookup(auth_settings_v2.value, "google_v2")] : []

        content {
          client_id                  = google_v2.value.client_id
          client_secret_setting_name = google_v2.value.client_secret_setting_name
          allowed_audiences          = google_v2.value.allowed_audiences
          login_scopes               = google_v2.value.login_scopes
        }
      }

      dynamic "microsoft_v2" {
        for_each = lookup(auth_settings_v2.value, "microsoft_v2", null) != null ? [lookup(auth_settings_v2.value, "microsoft_v2")] : []

        content {
          client_id                  = microsoft_v2.value.client_id
          client_secret_setting_name = microsoft_v2.value.client_secret_setting_name
          allowed_audiences          = microsoft_v2.value.allowed_audiences
          login_scopes               = microsoft_v2.value.login_scopes
        }
      }

      dynamic "twitter_v2" {
        for_each = lookup(auth_settings_v2.value, "twitter_v2", null) != null ? [lookup(auth_settings_v2.value, "twitter_v2")] : []

        content {
          consumer_key                 = twitter_v2.value.consumer_key
          consumer_secret_setting_name = twitter_v2.value.consumer_secret_setting_name
        }
      }
    }
  }

  dynamic "storage_account" {
    for_each = lookup(
      each.value, "storage_accounts", {}
    )

    content {
      name = lookup(
        storage_account.value, "name", storage_account.key
      )

      type         = storage_account.value.type
      share_name   = storage_account.value.share_name
      access_key   = storage_account.value.access_key
      account_name = storage_account.value.account_name
      mount_path   = storage_account.value.mount_path
    }
  }

  site_config {
    always_on                                     = var.instance.site_config.always_on
    ftps_state                                    = var.instance.site_config.ftps_state
    worker_count                                  = var.instance.site_config.worker_count
    http2_enabled                                 = var.instance.site_config.http2_enabled
    app_scale_limit                               = var.instance.site_config.app_scale_limit
    app_command_line                              = var.instance.site_config.app_command_line
    remote_debugging_version                      = var.instance.site_config.remote_debugging_version
    pre_warmed_instance_count                     = var.instance.site_config.pre_warmed_instance_count
    runtime_scale_monitoring_enabled              = var.instance.site_config.runtime_scale_monitoring_enabled
    scm_use_main_ip_restriction                   = var.instance.site_config.scm_use_main_ip_restriction
    health_check_eviction_time_in_min             = var.instance.site_config.health_check_eviction_time_in_min
    application_insights_connection_string        = var.instance.site_config.application_insights_connection_string
    container_registry_use_managed_identity       = var.instance.site_config.container_registry_use_managed_identity
    container_registry_managed_identity_client_id = var.instance.site_config.container_registry_managed_identity_client_id
    minimum_tls_version                           = var.instance.site_config.minimum_tls_version
    api_management_api_id                         = var.instance.site_config.api_management_api_id
    managed_pipeline_mode                         = var.instance.site_config.managed_pipeline_mode
    vnet_route_all_enabled                        = var.instance.site_config.vnet_route_all_enabled
    scm_minimum_tls_version                       = var.instance.site_config.scm_minimum_tls_version
    application_insights_key                      = var.instance.site_config.application_insights_key
    elastic_instance_minimum                      = var.instance.site_config.elastic_instance_minimum
    remote_debugging_enabled                      = var.instance.site_config.remote_debugging_enabled
    default_documents                             = var.instance.site_config.default_documents
    health_check_path                             = var.instance.site_config.health_check_path
    use_32_bit_worker                             = var.instance.site_config.use_32_bit_worker
    api_definition_url                            = var.instance.site_config.api_definition_url
    websockets_enabled                            = var.instance.site_config.websockets_enabled
    load_balancing_mode                           = var.instance.site_config.load_balancing_mode
    ip_restriction_default_action                 = var.instance.site_config.ip_restriction_default_action
    scm_ip_restriction_default_action             = var.instance.site_config.scm_ip_restriction_default_action

    dynamic "ip_restriction" {
      for_each = lookup(
        each.value, "ip_restrictions", {}
      )

      content {
        action                    = ip_restriction.value.action
        ip_address                = ip_restriction.value.ip_address
        name                      = ip_restriction.value.name
        priority                  = ip_restriction.value.priority
        service_tag               = ip_restriction.value.service_tag
        virtual_network_subnet_id = ip_restriction.value.virtual_network_subnet_id
        description               = ip_restriction.value.description
        headers                   = ip_restriction.value.headers
      }
    }

    dynamic "scm_ip_restriction" {
      for_each = lookup(
        each.value, "scm_ip_restrictions", {}
      )

      content {
        action                    = scm_ip_restriction.value.action
        ip_address                = scm_ip_restriction.value.ip_address
        name                      = scm_ip_restriction.value.name
        priority                  = scm_ip_restriction.value.priority
        service_tag               = scm_ip_restriction.value.service_tag
        virtual_network_subnet_id = scm_ip_restriction.value.virtual_network_subnet_id
        headers                   = scm_ip_restriction.value.headers
        description               = scm_ip_restriction.value.description
      }
    }

    dynamic "application_stack" {
      for_each = lookup(each.value.site_config, "application_stack", null) != null ? [lookup(each.value.site_config, "application_stack")] : []

      content {
        dotnet_version              = application_stack.value.dotnet_version
        use_dotnet_isolated_runtime = application_stack.value.use_dotnet_isolated_runtime
        java_version                = application_stack.value.java_version
        node_version                = application_stack.value.node_version
        python_version              = application_stack.value.python_version
        powershell_core_version     = application_stack.value.powershell_core_version
        use_custom_runtime          = application_stack.value.use_custom_runtime

        dynamic "docker" {
          for_each = lookup(application_stack.value, "docker", null) != null ? [lookup(application_stack.value, "docker")] : []
          content {
            image_name        = docker.value.image_name
            image_tag         = docker.value.image_tag
            registry_url      = docker.value.registry_url
            registry_username = docker.value.registry_username
            registry_password = docker.value.registry_password
          }
        }
      }
    }

    dynamic "cors" {
      for_each = lookup(each.value.site_config, "cors", null) != null ? [lookup(each.value.site_config, "cors")] : []

      content {
        allowed_origins     = cors.value.allowed_origins
        support_credentials = cors.value.support_credentials
      }
    }

    dynamic "app_service_logs" {
      for_each = lookup(each.value.site_config, "app_service_logs", null) != null ? [lookup(each.value.site_config, "app_service_logs")] : []

      content {
        disk_quota_mb         = app_service_logs.value.disk_quota_mb
        retention_period_days = app_service_logs.value.retention_period_days
      }
    }
  }

  dynamic "sticky_settings" {
    for_each = try(var.instance.sticky_settings, null) != null ? [1] : []

    content {
      app_setting_names       = var.instance.sticky_settings.app_setting_names
      connection_string_names = var.instance.sticky_settings.connection_string_names
    }
  }

  dynamic "backup" {
    for_each = try(var.instance.backup, null) != null ? [1] : []

    content {
      name                = var.instance.backup.name
      enabled             = var.instance.backup.enabled
      storage_account_url = var.instance.backup.storage_account_url

      schedule {
        frequency_unit           = var.instance.backup.schedule.frequency_unit
        frequency_interval       = var.instance.backup.schedule.frequency_interval
        retention_period_days    = var.instance.backup.schedule.retention_period_days
        start_time               = var.instance.backup.schedule.start_time
        keep_at_least_one_backup = var.instance.backup.schedule.keep_at_least_one_backup
      }
    }
  }

  dynamic "connection_string" {
    for_each = {
      for k, v in try(var.instance.connection_string, {}) : k => v
    }

    content {
      name  = connection_string.value.name
      type  = connection_string.value.type
      value = connection_string.value.value
    }
  }
  lifecycle {
    ignore_changes = [
      auth_settings,
      app_settings["WEBSITE_RUN_FROM_PACKAGE"],
      app_settings["WEBSITE_ENABLE_SYNC_UPDATE_SITE"],
      tags["hidden-link: /app-insights-instrumentation-key"],
      tags["hidden-link: /app-insights-resource-id"],
      tags["hidden-link: /app-insights-conn-string"],
    ]
  }
}

# linux function app slot
resource "azurerm_linux_function_app_slot" "slot" {
  for_each = {
    for key, value in try(var.instance.slots, {}) : key => value
    if var.instance.type == "linux"
  }

  name                                           = each.value.name
  function_app_id                                = var.instance.type == "linux" ? azurerm_linux_function_app.func[var.instance.name].id : azurerm_windows_function_app.func[var.instance.name].id
  storage_account_name                           = var.instance.storage_account_name
  storage_account_access_key                     = var.instance.storage_account_access_key
  webdeploy_publish_basic_authentication_enabled = var.instance.webdeploy_publish_basic_authentication_enabled
  ftp_publish_basic_authentication_enabled       = var.instance.ftp_publish_basic_authentication_enabled
  client_certificate_exclusion_paths             = var.instance.client_certificate_exclusion_paths
  key_vault_reference_identity_id                = var.instance.key_vault_reference_identity_id
  vnet_image_pull_enabled                        = var.instance.vnet_image_pull_enabled
  content_share_force_disabled                   = var.instance.content_share_force_disabled
  storage_uses_managed_identity                  = var.instance.storage_uses_managed_identity
  enabled                                        = var.instance.enabled
  public_network_access_enabled                  = var.instance.public_network_access_enabled
  storage_key_vault_secret_id                    = var.instance.storage_key_vault_secret_id
  functions_extension_version                    = var.instance.functions_extension_version
  client_certificate_enabled                     = var.instance.client_certificate_enabled
  virtual_network_subnet_id                      = var.instance.virtual_network_subnet_id
  daily_memory_time_quota                        = var.instance.daily_memory_time_quota
  client_certificate_mode                        = var.instance.client_certificate_mode
  builtin_logging_enabled                        = var.instance.builtin_logging_enabled
  https_only                                     = var.instance.https_only
  virtual_network_backup_restore_enabled         = var.instance.virtual_network_backup_restore_enabled
  app_settings                                   = each.value.app_settings

  service_plan_id = lookup(
    each.value, "service_plan_id"
  )

  tags = try(
    var.instance.tags, var.tags, null
  )


  dynamic "identity" {
    for_each = lookup(var.instance, "identity", null) != null ? [var.instance.identity] : []

    content {
      type         = identity.value.type
      identity_ids = identity.value.identity_ids
    }
  }

  dynamic "auth_settings_v2" {
    for_each = lookup(each.value, "auth_settings_v2", null) != null ? [lookup(each.value, "auth_settings_v2")] : []

    content {
      auth_enabled                            = auth_settings_v2.value.auth_enabled
      runtime_version                         = auth_settings_v2.value.runtime_version
      config_file_path                        = auth_settings_v2.value.config_file_path
      require_authentication                  = auth_settings_v2.value.require_authentication
      unauthenticated_action                  = auth_settings_v2.value.unauthenticated_action
      default_provider                        = auth_settings_v2.value.default_provider
      excluded_paths                          = auth_settings_v2.value.excluded_paths
      require_https                           = auth_settings_v2.value.require_https
      http_route_api_prefix                   = auth_settings_v2.value.http_route_api_prefix
      forward_proxy_convention                = auth_settings_v2.value.forward_proxy_convention
      forward_proxy_custom_host_header_name   = auth_settings_v2.value.forward_proxy_custom_host_header_name
      forward_proxy_custom_scheme_header_name = auth_settings_v2.value.forward_proxy_custom_scheme_header_name

      login {
        token_store_enabled               = auth_settings_v2.value.login.token_store_enabled
        token_refresh_extension_time      = auth_settings_v2.value.login.token_refresh_extension_time
        token_store_path                  = auth_settings_v2.value.login.token_store_path
        token_store_sas_setting_name      = auth_settings_v2.value.login.token_store_sas_setting_name
        preserve_url_fragments_for_logins = auth_settings_v2.value.login.preserve_url_fragments_for_logins
        allowed_external_redirect_urls    = auth_settings_v2.value.login.allowed_external_redirect_urls
        cookie_expiration_convention      = auth_settings_v2.value.login.cookie_expiration_convention
        cookie_expiration_time            = auth_settings_v2.value.login.cookie_expiration_time
        validate_nonce                    = auth_settings_v2.value.login.validate_nonce
        nonce_expiration_time             = auth_settings_v2.value.login.nonce_expiration_time
        logout_endpoint                   = auth_settings_v2.value.login.logout_endpoint
      }

      dynamic "apple_v2" {
        for_each = lookup(auth_settings_v2.value, "apple_v2", null) != null ? [lookup(auth_settings_v2.value, "apple_v2")] : []

        content {
          client_id                  = apple_v2.value.client_id
          client_secret_setting_name = apple_v2.value.client_secret_setting_name
          login_scopes               = apple_v2.value.login_scopes
        }
      }

      dynamic "active_directory_v2" {
        for_each = lookup(auth_settings_v2.value, "active_directory_v2", null) != null ? [lookup(auth_settings_v2.value, "active_directory_v2")] : []

        content {
          client_id                            = active_directory_v2.value.client_id
          tenant_auth_endpoint                 = active_directory_v2.value.tenant_auth_endpoint
          client_secret_setting_name           = active_directory_v2.value.client_secret_setting_name
          client_secret_certificate_thumbprint = active_directory_v2.value.client_secret_certificate_thumbprint
          jwt_allowed_groups                   = active_directory_v2.value.jwt_allowed_groups
          jwt_allowed_client_applications      = active_directory_v2.value.jwt_allowed_client_applications
          www_authentication_disabled          = active_directory_v2.value.www_authentication_disabled
          allowed_audiences                    = active_directory_v2.value.allowed_audiences
          allowed_groups                       = active_directory_v2.value.allowed_groups
          allowed_identities                   = active_directory_v2.value.allowed_identities
          login_parameters                     = active_directory_v2.value.login_parameters
          allowed_applications                 = active_directory_v2.value.allowed_applications
        }
      }

      dynamic "azure_static_web_app_v2" {
        for_each = lookup(auth_settings_v2.value, "azure_static_web_app_v2", null) != null ? [lookup(auth_settings_v2.value, "azure_static_web_app_v2")] : []

        content {
          client_id = azure_static_web_app_v2.value.client_id
        }
      }

      dynamic "custom_oidc_v2" {
        for_each = try(auth_settings_v2.value.custom_oidc_v2, {})

        content {
          name                          = custom_oidc_v2.value.name
          client_id                     = custom_oidc_v2.value.client_id
          openid_configuration_endpoint = custom_oidc_v2.value.openid_configuration_endpoint
          name_claim_type               = custom_oidc_v2.value.name_claim_type
          scopes                        = custom_oidc_v2.value.scopes
          client_credential_method      = custom_oidc_v2.value.client_credential_method
          client_secret_setting_name    = custom_oidc_v2.value.client_secret_setting_name
          authorisation_endpoint        = custom_oidc_v2.value.authorisation_endpoint
          token_endpoint                = custom_oidc_v2.value.token_endpoint
          issuer_endpoint               = custom_oidc_v2.value.issuer_endpoint
          certification_uri             = custom_oidc_v2.value.certification_uri
        }
      }

      dynamic "facebook_v2" {
        for_each = lookup(auth_settings_v2.value, "facebook_v2", null) != null ? [lookup(auth_settings_v2.value, "facebook_v2")] : []

        content {
          app_id                  = facebook_v2.value.app_id
          app_secret_setting_name = facebook_v2.value.app_secret_setting_name
          graph_api_version       = facebook_v2.value.graph_api_version
          login_scopes            = facebook_v2.value.login_scopes
        }
      }

      dynamic "github_v2" {
        for_each = lookup(auth_settings_v2.value, "github_v2", null) != null ? [lookup(auth_settings_v2.value, "github_v2")] : []

        content {
          client_id                  = github_v2.value.client_id
          client_secret_setting_name = github_v2.value.client_secret_setting_name
          login_scopes               = github_v2.value.login_scopes
        }
      }

      dynamic "google_v2" {
        for_each = lookup(auth_settings_v2.value, "google_v2", null) != null ? [lookup(auth_settings_v2.value, "google_v2")] : []

        content {
          client_id                  = google_v2.value.client_id
          client_secret_setting_name = google_v2.value.client_secret_setting_name
          allowed_audiences          = google_v2.value.allowed_audiences
          login_scopes               = google_v2.value.login_scopes
        }
      }

      dynamic "microsoft_v2" {
        for_each = lookup(auth_settings_v2.value, "microsoft_v2", null) != null ? [lookup(auth_settings_v2.value, "microsoft_v2")] : []

        content {
          client_id                  = microsoft_v2.value.client_id
          client_secret_setting_name = microsoft_v2.value.client_secret_setting_name
          allowed_audiences          = microsoft_v2.value.allowed_audiences
          login_scopes               = microsoft_v2.value.login_scopes
        }
      }

      dynamic "twitter_v2" {
        for_each = lookup(auth_settings_v2.value, "twitter_v2", null) != null ? [lookup(auth_settings_v2.value, "twitter_v2")] : []

        content {
          consumer_key                 = twitter_v2.value.consumer_key
          consumer_secret_setting_name = twitter_v2.value.consumer_secret_setting_name
        }
      }
    }
  }

  dynamic "storage_account" {
    for_each = lookup(
      each.value, "storage_accounts", {}
    )

    content {
      name = lookup(
        storage_account.value, "name", storage_account.key
      )

      type         = storage_account.value.type
      share_name   = storage_account.value.share_name
      access_key   = storage_account.value.access_key
      account_name = storage_account.value.account_name
      mount_path   = storage_account.value.mount_path
    }
  }

  dynamic "connection_string" {
    for_each = lookup(
      each.value, "connection_strings", {}
    )

    content {
      name = lookup(
        connection_string.value, "name", connection_string.key
      )

      value = connection_string.value
      type  = connection_string.type
    }
  }

  dynamic "backup" {
    for_each = lookup(each.value, "backup", null) != null ? [lookup(each.value, "backup")] : []

    content {
      name                = backup.value.name
      storage_account_url = backup.value.storage_account_url
      enabled             = backup.value.enabled

      schedule {
        frequency_interval       = backup.value.schedule.frequency_interval
        frequency_unit           = backup.value.schedule.frequency_unit
        keep_at_least_one_backup = backup.value.schedule.keep_at_least_one_backup
        retention_period_days    = backup.value.schedule.retention_period_days
        start_time               = backup.value.schedule.start_time
        last_execution_time      = backup.value.schedule.last_execution_time
      }
    }
  }

  site_config {
    always_on                                     = each.value.site_config.always_on
    ftps_state                                    = each.value.site_config.ftps_state
    worker_count                                  = each.value.site_config.worker_count
    http2_enabled                                 = each.value.site_config.http2_enabled
    app_scale_limit                               = each.value.site_config.app_scale_limit
    app_command_line                              = each.value.site_config.app_command_line
    remote_debugging_version                      = each.value.site_config.remote_debugging_version
    pre_warmed_instance_count                     = each.value.site_config.pre_warmed_instance_count
    runtime_scale_monitoring_enabled              = each.value.site_config.runtime_scale_monitoring_enabled
    scm_use_main_ip_restriction                   = each.value.site_config.scm_use_main_ip_restriction
    health_check_eviction_time_in_min             = each.value.site_config.health_check_eviction_time_in_min
    application_insights_connection_string        = each.value.site_config.application_insights_connection_string
    container_registry_use_managed_identity       = each.value.site_config.container_registry_use_managed_identity
    container_registry_managed_identity_client_id = each.value.site_config.container_registry_managed_identity_client_id
    minimum_tls_version                           = each.value.site_config.minimum_tls_version
    api_management_api_id                         = each.value.site_config.api_management_api_id
    managed_pipeline_mode                         = each.value.site_config.managed_pipeline_mode
    vnet_route_all_enabled                        = each.value.site_config.vnet_route_all_enabled
    scm_minimum_tls_version                       = each.value.site_config.scm_minimum_tls_version
    application_insights_key                      = each.value.site_config.application_insights_key
    elastic_instance_minimum                      = each.value.site_config.elastic_instance_minimum
    remote_debugging_enabled                      = each.value.site_config.remote_debugging_enabled
    default_documents                             = each.value.site_config.default_documents
    health_check_path                             = each.value.site_config.health_check_path
    use_32_bit_worker                             = each.value.site_config.use_32_bit_worker
    api_definition_url                            = each.value.site_config.api_definition_url
    auto_swap_slot_name                           = each.value.site_config.auto_swap_slot_name
    websockets_enabled                            = each.value.site_config.websockets_enabled
    load_balancing_mode                           = each.value.site_config.load_balancing_mode
    scm_ip_restriction_default_action             = each.value.site_config.scm_ip_restriction_default_action
    ip_restriction_default_action                 = each.value.site_config.ip_restriction_default_action

    dynamic "ip_restriction" {
      for_each = try(
        each.value.site_config.ip_restrictions, {}
      )

      content {
        action                    = ip_restriction.value.action
        ip_address                = ip_restriction.value.ip_address
        name                      = ip_restriction.value.name
        priority                  = ip_restriction.value.priority
        service_tag               = ip_restriction.value.service_tag
        virtual_network_subnet_id = ip_restriction.value.virtual_network_subnet_id
        description               = ip_restriction.value.description
        headers                   = ip_restriction.value.headers
      }
    }

    dynamic "scm_ip_restriction" {
      for_each = try(
        var.instance.site_config.scm_ip_restrictions, {}
      )

      content {
        action                    = scm_ip_restriction.value.action
        ip_address                = scm_ip_restriction.value.ip_address
        name                      = scm_ip_restriction.value.name
        priority                  = scm_ip_restriction.value.priority
        service_tag               = scm_ip_restriction.value.service_tag
        virtual_network_subnet_id = scm_ip_restriction.value.virtual_network_subnet_id
        headers                   = scm_ip_restriction.value.headers
        description               = scm_ip_restriction.value.description
      }
    }

    dynamic "application_stack" {
      for_each = lookup(each.value.site_config, "application_stack", null) != null ? [lookup(each.value.site_config, "application_stack")] : []

      content {
        dotnet_version              = application_stack.value.dotnet_version
        use_dotnet_isolated_runtime = application_stack.value.use_dotnet_isolated_runtime
        java_version                = application_stack.value.java_version
        node_version                = application_stack.value.node_version
        python_version              = application_stack.value.python_version
        powershell_core_version     = application_stack.value.powershell_core_version
        use_custom_runtime          = application_stack.value.use_custom_runtime

        dynamic "docker" {
          for_each = lookup(application_stack.value, "docker", null) != null ? [lookup(application_stack.value, "docker")] : []

          content {
            image_name        = docker.value.image_name
            image_tag         = docker.value.image_tag
            registry_url      = docker.value.registry_url
            registry_username = docker.value.registry_username
            registry_password = docker.value.registry_password
          }
        }
      }
    }

    dynamic "cors" {
      for_each = lookup(each.value.site_config, "cors", null) != null ? [lookup(each.value.site_config, "cors")] : []

      content {
        allowed_origins     = cors.value.allowed_origins
        support_credentials = cors.value.support_credentials
      }
    }

    dynamic "app_service_logs" {
      for_each = lookup(each.value.site_config, "app_service_logs", null) != null ? [lookup(each.value.site_config, "app_service_logs")] : []

      content {
        disk_quota_mb         = app_service_logs.value.disk_quota_mb
        retention_period_days = app_service_logs.value.retention_period_days
      }
    }
  }
  lifecycle {
    ignore_changes = [
      auth_settings,
      app_settings["WEBSITE_RUN_FROM_PACKAGE"],
      app_settings["WEBSITE_ENABLE_SYNC_UPDATE_SITE"],
      tags["hidden-link: /app-insights-instrumentation-key"],
      tags["hidden-link: /app-insights-resource-id"],
      tags["hidden-link: /app-insights-conn-string"],
    ]
  }
}

# windows function app
resource "azurerm_windows_function_app" "func" {
  for_each = var.instance.type == "windows" ? {
    (var.instance.name) = var.instance
  } : {}

  resource_group_name = coalesce(
    lookup(
      var.instance, "resource_group", null
    ), var.resource_group
  )

  location = coalesce(
    lookup(var.instance, "location", null
    ), var.location
  )

  name                                           = var.instance.name
  service_plan_id                                = var.instance.service_plan_id
  storage_account_name                           = var.instance.storage_account_name
  storage_account_access_key                     = var.instance.storage_account_access_key
  https_only                                     = var.instance.https_only
  zip_deploy_file                                = var.instance.zip_deploy_file
  enabled                                        = var.instance.enabled
  builtin_logging_enabled                        = var.instance.builtin_logging_enabled
  client_certificate_mode                        = var.instance.client_certificate_mode
  daily_memory_time_quota                        = var.instance.daily_memory_time_quota
  virtual_network_subnet_id                      = var.instance.virtual_network_subnet_id
  client_certificate_enabled                     = var.instance.client_certificate_enabled
  functions_extension_version                    = var.instance.functions_extension_version
  storage_key_vault_secret_id                    = var.instance.storage_key_vault_secret_id
  content_share_force_disabled                   = var.instance.content_share_force_disabled
  public_network_access_enabled                  = var.instance.public_network_access_enabled
  storage_uses_managed_identity                  = var.instance.storage_uses_managed_identity
  vnet_image_pull_enabled                        = var.instance.vnet_image_pull_enabled
  key_vault_reference_identity_id                = var.instance.key_vault_reference_identity_id
  client_certificate_exclusion_paths             = var.instance.client_certificate_exclusion_paths
  ftp_publish_basic_authentication_enabled       = var.instance.ftp_publish_basic_authentication_enabled
  webdeploy_publish_basic_authentication_enabled = var.instance.webdeploy_publish_basic_authentication_enabled
  virtual_network_backup_restore_enabled         = var.instance.virtual_network_backup_restore_enabled
  app_settings                                   = var.instance.app_settings

  tags = try(
    var.instance.tags, var.tags, null
  )

  dynamic "identity" {
    for_each = lookup(var.instance, "identity", null) != null ? [var.instance.identity] : []

    content {
      type         = identity.value.type
      identity_ids = identity.value.identity_ids
    }
  }

  dynamic "auth_settings_v2" {
    for_each = lookup(each.value, "auth_settings_v2", null) != null ? [lookup(each.value, "auth_settings_v2")] : []

    content {
      auth_enabled                            = auth_settings_v2.value.auth_enabled
      runtime_version                         = auth_settings_v2.value.runtime_version
      config_file_path                        = auth_settings_v2.value.config_file_path
      require_authentication                  = auth_settings_v2.value.require_authentication
      unauthenticated_action                  = auth_settings_v2.value.unauthenticated_action
      default_provider                        = auth_settings_v2.value.default_provider
      excluded_paths                          = auth_settings_v2.value.excluded_paths
      require_https                           = auth_settings_v2.value.require_https
      http_route_api_prefix                   = auth_settings_v2.value.http_route_api_prefix
      forward_proxy_convention                = auth_settings_v2.value.forward_proxy_convention
      forward_proxy_custom_host_header_name   = auth_settings_v2.value.forward_proxy_custom_host_header_name
      forward_proxy_custom_scheme_header_name = auth_settings_v2.value.forward_proxy_custom_scheme_header_name

      login {
        token_store_enabled               = auth_settings_v2.value.login.token_store_enabled
        token_refresh_extension_time      = auth_settings_v2.value.login.token_refresh_extension_time
        token_store_path                  = auth_settings_v2.value.login.token_store_path
        token_store_sas_setting_name      = auth_settings_v2.value.login.token_store_sas_setting_name
        preserve_url_fragments_for_logins = auth_settings_v2.value.login.preserve_url_fragments_for_logins
        allowed_external_redirect_urls    = auth_settings_v2.value.login.allowed_external_redirect_urls
        cookie_expiration_convention      = auth_settings_v2.value.login.cookie_expiration_convention
        cookie_expiration_time            = auth_settings_v2.value.login.cookie_expiration_time
        validate_nonce                    = auth_settings_v2.value.login.validate_nonce
        nonce_expiration_time             = auth_settings_v2.value.login.nonce_expiration_time
        logout_endpoint                   = auth_settings_v2.value.login.logout_endpoint
      }

      dynamic "apple_v2" {
        for_each = lookup(auth_settings_v2.value, "apple_v2", null) != null ? [lookup(auth_settings_v2.value, "apple_v2")] : []

        content {
          client_id                  = apple_v2.value.client_id
          client_secret_setting_name = apple_v2.value.client_secret_setting_name
          login_scopes               = apple_v2.value.login_scopes
        }
      }

      dynamic "active_directory_v2" {
        for_each = lookup(auth_settings_v2.value, "active_directory_v2", null) != null ? [lookup(auth_settings_v2.value, "active_directory_v2")] : []

        content {
          client_id                            = active_directory_v2.value.client_id
          tenant_auth_endpoint                 = active_directory_v2.value.tenant_auth_endpoint
          client_secret_setting_name           = active_directory_v2.value.client_secret_setting_name
          client_secret_certificate_thumbprint = active_directory_v2.value.client_secret_certificate_thumbprint
          jwt_allowed_groups                   = active_directory_v2.value.jwt_allowed_groups
          jwt_allowed_client_applications      = active_directory_v2.value.jwt_allowed_client_applications
          www_authentication_disabled          = active_directory_v2.value.www_authentication_disabled
          allowed_audiences                    = active_directory_v2.value.allowed_audiences
          allowed_groups                       = active_directory_v2.value.allowed_groups
          allowed_identities                   = active_directory_v2.value.allowed_identities
          login_parameters                     = active_directory_v2.value.login_parameters
          allowed_applications                 = active_directory_v2.value.allowed_applications
        }
      }

      dynamic "azure_static_web_app_v2" {
        for_each = lookup(auth_settings_v2.value, "azure_static_web_app_v2", null) != null ? [lookup(auth_settings_v2.value, "azure_static_web_app_v2")] : []

        content {
          client_id = azure_static_web_app_v2.value.client_id
        }
      }

      dynamic "custom_oidc_v2" {
        for_each = try(auth_settings_v2.value.custom_oidc_v2, {})

        content {
          name                          = custom_oidc_v2.value.name
          client_id                     = custom_oidc_v2.value.client_id
          openid_configuration_endpoint = custom_oidc_v2.value.openid_configuration_endpoint
          name_claim_type               = custom_oidc_v2.value.name_claim_type
          scopes                        = custom_oidc_v2.value.scopes
          client_credential_method      = custom_oidc_v2.value.client_credential_method
          client_secret_setting_name    = custom_oidc_v2.value.client_secret_setting_name
          authorisation_endpoint        = custom_oidc_v2.value.authorisation_endpoint
          token_endpoint                = custom_oidc_v2.value.token_endpoint
          issuer_endpoint               = custom_oidc_v2.value.issuer_endpoint
          certification_uri             = custom_oidc_v2.value.certification_uri
        }
      }

      dynamic "facebook_v2" {
        for_each = lookup(auth_settings_v2.value, "facebook_v2", null) != null ? [lookup(auth_settings_v2.value, "facebook_v2")] : []

        content {
          app_id                  = facebook_v2.value.app_id
          app_secret_setting_name = facebook_v2.value.app_secret_setting_name
          graph_api_version       = facebook_v2.value.graph_api_version
          login_scopes            = facebook_v2.value.login_scopes
        }
      }

      dynamic "github_v2" {
        for_each = lookup(auth_settings_v2.value, "github_v2", null) != null ? [lookup(auth_settings_v2.value, "github_v2")] : []

        content {
          client_id                  = github_v2.value.client_id
          client_secret_setting_name = github_v2.value.client_secret_setting_name
          login_scopes               = github_v2.value.login_scopes
        }
      }

      dynamic "google_v2" {
        for_each = lookup(auth_settings_v2.value, "google_v2", null) != null ? [lookup(auth_settings_v2.value, "google_v2")] : []

        content {
          client_id                  = google_v2.value.client_id
          client_secret_setting_name = google_v2.value.client_secret_setting_name
          allowed_audiences          = google_v2.value.allowed_audiences
          login_scopes               = google_v2.value.login_scopes
        }
      }

      dynamic "microsoft_v2" {
        for_each = lookup(auth_settings_v2.value, "microsoft_v2", null) != null ? [lookup(auth_settings_v2.value, "microsoft_v2")] : []

        content {
          client_id                  = microsoft_v2.value.client_id
          client_secret_setting_name = microsoft_v2.value.client_secret_setting_name
          allowed_audiences          = microsoft_v2.value.allowed_audiences
          login_scopes               = microsoft_v2.value.login_scopes
        }
      }

      dynamic "twitter_v2" {
        for_each = lookup(auth_settings_v2.value, "twitter_v2", null) != null ? [lookup(auth_settings_v2.value, "twitter_v2")] : []

        content {
          consumer_key                 = twitter_v2.value.consumer_key
          consumer_secret_setting_name = twitter_v2.value.consumer_secret_setting_name
        }
      }
    }
  }

  dynamic "storage_account" {
    for_each = lookup(
      each.value, "storage_accounts", {}
    )

    content {
      name = lookup(
        storage_account.value, "name", storage_account.key
      )

      type         = storage_account.value.type
      share_name   = storage_account.value.share_name
      access_key   = storage_account.value.access_key
      account_name = storage_account.value.account_name
      mount_path   = storage_account.value.mount_path
    }
  }

  site_config {
    always_on                              = var.instance.site_config.always_on
    ftps_state                             = var.instance.site_config.ftps_state
    worker_count                           = var.instance.site_config.worker_count
    http2_enabled                          = var.instance.site_config.http2_enabled
    app_scale_limit                        = var.instance.site_config.app_scale_limit
    app_command_line                       = var.instance.site_config.app_command_line
    remote_debugging_version               = var.instance.site_config.remote_debugging_version
    pre_warmed_instance_count              = var.instance.site_config.pre_warmed_instance_count
    runtime_scale_monitoring_enabled       = var.instance.site_config.runtime_scale_monitoring_enabled
    scm_use_main_ip_restriction            = var.instance.site_config.scm_use_main_ip_restriction
    health_check_eviction_time_in_min      = var.instance.site_config.health_check_eviction_time_in_min
    application_insights_connection_string = var.instance.site_config.application_insights_connection_string
    minimum_tls_version                    = var.instance.site_config.minimum_tls_version
    api_management_api_id                  = var.instance.site_config.api_management_api_id
    managed_pipeline_mode                  = var.instance.site_config.managed_pipeline_mode
    vnet_route_all_enabled                 = var.instance.site_config.vnet_route_all_enabled
    scm_minimum_tls_version                = var.instance.site_config.scm_minimum_tls_version
    application_insights_key               = var.instance.site_config.application_insights_key
    elastic_instance_minimum               = var.instance.site_config.elastic_instance_minimum
    remote_debugging_enabled               = var.instance.site_config.remote_debugging_enabled
    default_documents                      = var.instance.site_config.default_documents
    health_check_path                      = var.instance.site_config.health_check_path
    use_32_bit_worker                      = var.instance.site_config.use_32_bit_worker
    api_definition_url                     = var.instance.site_config.api_definition_url
    websockets_enabled                     = var.instance.site_config.websockets_enabled
    load_balancing_mode                    = var.instance.site_config.load_balancing_mode
    ip_restriction_default_action          = var.instance.site_config.ip_restriction_default_action
    scm_ip_restriction_default_action      = var.instance.site_config.scm_ip_restriction_default_action

    dynamic "ip_restriction" {
      for_each = lookup(
        each.value, "ip_restrictions", {}
      )

      content {
        action                    = ip_restriction.value.action
        ip_address                = ip_restriction.value.ip_address
        name                      = ip_restriction.value.name
        priority                  = ip_restriction.value.priority
        service_tag               = ip_restriction.value.service_tag
        virtual_network_subnet_id = ip_restriction.value.virtual_network_subnet_id
        description               = ip_restriction.value.description
        headers                   = ip_restriction.value.headers
      }
    }

    dynamic "scm_ip_restriction" {
      for_each = lookup(
        each.value, "scm_ip_restrictions", {}
      )

      content {
        action                    = scm_ip_restriction.value.action
        ip_address                = scm_ip_restriction.value.ip_address
        name                      = scm_ip_restriction.value.name
        priority                  = scm_ip_restriction.value.priority
        service_tag               = scm_ip_restriction.value.service_tag
        virtual_network_subnet_id = scm_ip_restriction.value.virtual_network_subnet_id
        headers                   = scm_ip_restriction.value.headers
        description               = scm_ip_restriction.value.description
      }
    }

    dynamic "application_stack" {
      for_each = try(var.instance.site_config.application_stack, null) != null ? [1] : []

      content {
        dotnet_version              = var.instance.site_config.application_stack.dotnet_version
        use_dotnet_isolated_runtime = var.instance.site_config.application_stack.use_dotnet_isolated_runtime
        java_version                = var.instance.site_config.application_stack.java_version
        node_version                = var.instance.site_config.application_stack.node_version
        powershell_core_version     = var.instance.site_config.application_stack.powershell_core_version
        use_custom_runtime          = var.instance.site_config.application_stack.use_custom_runtime
      }
    }

    dynamic "cors" {
      for_each = try(var.instance.site_config.cors, null) != null ? [1] : []

      content {
        allowed_origins     = var.instance.site_config.cors.allowed_origins
        support_credentials = var.instance.site_config.cors.support_credentials
      }
    }

    dynamic "app_service_logs" {
      for_each = try(var.instance.site_config.app_service_logs, null) != null ? [1] : []

      content {
        disk_quota_mb         = var.instance.site_config.app_service_logs.disk_quota_mb
        retention_period_days = var.instance.site_config.app_service_logs.retention_period_days
      }
    }
  }

  dynamic "sticky_settings" {
    for_each = try(var.instance.sticky_settings, null) != null ? [1] : []

    content {
      app_setting_names       = var.instance.sticky_settings.app_setting_names
      connection_string_names = var.instance.sticky_settings.connection_string_names
    }
  }

  dynamic "backup" {
    for_each = try(var.instance.backup, null) != null ? [1] : []

    content {
      name                = var.instance.backup.name
      enabled             = var.instance.backup.enabled
      storage_account_url = var.instance.backup.storage_account_url

      schedule {
        frequency_unit           = var.instance.backup.schedule.frequency_unit
        frequency_interval       = var.instance.backup.schedule.frequency_interval
        retention_period_days    = var.instance.backup.schedule.retention_period_days
        start_time               = var.instance.backup.schedule.start_time
        keep_at_least_one_backup = var.instance.backup.schedule.keep_at_least_one_backup
      }
    }
  }

  dynamic "connection_string" {
    for_each = {
      for k, v in try(var.instance.connection_string, {}) : k => v
    }

    content {
      name  = connection_string.value.name
      type  = connection_string.value.type
      value = connection_string.value.value
    }
  }
  lifecycle {
    ignore_changes = [
      auth_settings,
      app_settings["WEBSITE_RUN_FROM_PACKAGE"],
      app_settings["WEBSITE_ENABLE_SYNC_UPDATE_SITE"],
      tags["hidden-link: /app-insights-instrumentation-key"],
      tags["hidden-link: /app-insights-resource-id"],
      tags["hidden-link: /app-insights-conn-string"],
    ]
  }
}

# windows function app slot
resource "azurerm_windows_function_app_slot" "slot" {
  for_each = {
    for key, value in try(var.instance.slots, {}) : key => value
    if var.instance.type == "windows"
  }

  name            = each.value.name
  function_app_id = var.instance.type == "linux" ? azurerm_linux_function_app.func[var.instance.name].id : azurerm_windows_function_app.func[var.instance.name].id

  storage_account_name                           = var.instance.storage_account_name
  storage_account_access_key                     = var.instance.storage_account_access_key
  webdeploy_publish_basic_authentication_enabled = var.instance.webdeploy_publish_basic_authentication_enabled
  ftp_publish_basic_authentication_enabled       = var.instance.ftp_publish_basic_authentication_enabled
  storage_uses_managed_identity                  = var.instance.storage_uses_managed_identity
  public_network_access_enabled                  = var.instance.public_network_access_enabled
  content_share_force_disabled                   = var.instance.content_share_force_disabled
  storage_key_vault_secret_id                    = var.instance.storage_key_vault_secret_id
  functions_extension_version                    = var.instance.functions_extension_version
  client_certificate_enabled                     = var.instance.client_certificate_enabled
  virtual_network_subnet_id                      = var.instance.virtual_network_subnet_id
  client_certificate_exclusion_paths             = var.instance.client_certificate_exclusion_paths
  key_vault_reference_identity_id                = var.instance.key_vault_reference_identity_id
  vnet_image_pull_enabled                        = var.instance.vnet_image_pull_enabled
  daily_memory_time_quota                        = var.instance.daily_memory_time_quota
  client_certificate_mode                        = var.instance.client_certificate_mode
  builtin_logging_enabled                        = var.instance.builtin_logging_enabled
  https_only                                     = var.instance.https_only
  enabled                                        = var.instance.enabled
  virtual_network_backup_restore_enabled         = var.instance.virtual_network_backup_restore_enabled
  app_settings                                   = each.value.app_settings

  service_plan_id = lookup(
    each.value, "service_plan_id"
  )

  tags = try(
    var.instance.tags, var.tags, null
  )

  dynamic "identity" {
    for_each = lookup(var.instance, "identity", null) != null ? [var.instance.identity] : []

    content {
      type         = identity.value.type
      identity_ids = identity.value.identity_ids
    }
  }

  dynamic "auth_settings_v2" {
    for_each = lookup(each.value, "auth_settings_v2", null) != null ? [lookup(each.value, "auth_settings_v2")] : []

    content {
      auth_enabled                            = auth_settings_v2.value.auth_enabled
      runtime_version                         = auth_settings_v2.value.runtime_version
      config_file_path                        = auth_settings_v2.value.config_file_path
      require_authentication                  = auth_settings_v2.value.require_authentication
      unauthenticated_action                  = auth_settings_v2.value.unauthenticated_action
      default_provider                        = auth_settings_v2.value.default_provider
      excluded_paths                          = auth_settings_v2.value.excluded_paths
      require_https                           = auth_settings_v2.value.require_https
      http_route_api_prefix                   = auth_settings_v2.value.http_route_api_prefix
      forward_proxy_convention                = auth_settings_v2.value.forward_proxy_convention
      forward_proxy_custom_host_header_name   = auth_settings_v2.value.forward_proxy_custom_host_header_name
      forward_proxy_custom_scheme_header_name = auth_settings_v2.value.forward_proxy_custom_scheme_header_name

      login {
        token_store_enabled               = auth_settings_v2.value.login.token_store_enabled
        token_refresh_extension_time      = auth_settings_v2.value.login.token_refresh_extension_time
        token_store_path                  = auth_settings_v2.value.login.token_store_path
        token_store_sas_setting_name      = auth_settings_v2.value.login.token_store_sas_setting_name
        preserve_url_fragments_for_logins = auth_settings_v2.value.login.preserve_url_fragments_for_logins
        allowed_external_redirect_urls    = auth_settings_v2.value.login.allowed_external_redirect_urls
        cookie_expiration_convention      = auth_settings_v2.value.login.cookie_expiration_convention
        cookie_expiration_time            = auth_settings_v2.value.login.cookie_expiration_time
        validate_nonce                    = auth_settings_v2.value.login.validate_nonce
        nonce_expiration_time             = auth_settings_v2.value.login.nonce_expiration_time
        logout_endpoint                   = auth_settings_v2.value.login.logout_endpoint
      }

      dynamic "apple_v2" {
        for_each = lookup(auth_settings_v2.value, "apple_v2", null) != null ? [lookup(auth_settings_v2.value, "apple_v2")] : []

        content {
          client_id                  = apple_v2.value.client_id
          client_secret_setting_name = apple_v2.value.client_secret_setting_name
          login_scopes               = apple_v2.value.login_scopes
        }
      }

      dynamic "active_directory_v2" {
        for_each = lookup(auth_settings_v2.value, "active_directory_v2", null) != null ? [lookup(auth_settings_v2.value, "active_directory_v2")] : []

        content {
          client_id                            = active_directory_v2.value.client_id
          tenant_auth_endpoint                 = active_directory_v2.value.tenant_auth_endpoint
          client_secret_setting_name           = active_directory_v2.value.client_secret_setting_name
          client_secret_certificate_thumbprint = active_directory_v2.value.client_secret_certificate_thumbprint
          jwt_allowed_groups                   = active_directory_v2.value.jwt_allowed_groups
          jwt_allowed_client_applications      = active_directory_v2.value.jwt_allowed_client_applications
          www_authentication_disabled          = active_directory_v2.value.www_authentication_disabled
          allowed_audiences                    = active_directory_v2.value.allowed_audiences
          allowed_groups                       = active_directory_v2.value.allowed_groups
          allowed_identities                   = active_directory_v2.value.allowed_identities
          login_parameters                     = active_directory_v2.value.login_parameters
          allowed_applications                 = active_directory_v2.value.allowed_applications
        }
      }

      dynamic "azure_static_web_app_v2" {
        for_each = lookup(auth_settings_v2.value, "azure_static_web_app_v2", null) != null ? [lookup(auth_settings_v2.value, "azure_static_web_app_v2")] : []

        content {
          client_id = azure_static_web_app_v2.value.client_id
        }
      }

      dynamic "custom_oidc_v2" {
        for_each = try(auth_settings_v2.value.custom_oidc_v2, {})

        content {
          name                          = custom_oidc_v2.value.name
          client_id                     = custom_oidc_v2.value.client_id
          openid_configuration_endpoint = custom_oidc_v2.value.openid_configuration_endpoint
          name_claim_type               = custom_oidc_v2.value.name_claim_type
          scopes                        = custom_oidc_v2.value.scopes
          client_credential_method      = custom_oidc_v2.value.client_credential_method
          client_secret_setting_name    = custom_oidc_v2.value.client_secret_setting_name
          authorisation_endpoint        = custom_oidc_v2.value.authorisation_endpoint
          token_endpoint                = custom_oidc_v2.value.token_endpoint
          issuer_endpoint               = custom_oidc_v2.value.issuer_endpoint
          certification_uri             = custom_oidc_v2.value.certification_uri
        }
      }

      dynamic "facebook_v2" {
        for_each = lookup(auth_settings_v2.value, "facebook_v2", null) != null ? [lookup(auth_settings_v2.value, "facebook_v2")] : []

        content {
          app_id                  = facebook_v2.value.app_id
          app_secret_setting_name = facebook_v2.value.app_secret_setting_name
          graph_api_version       = facebook_v2.value.graph_api_version
          login_scopes            = facebook_v2.value.login_scopes
        }
      }

      dynamic "github_v2" {
        for_each = lookup(auth_settings_v2.value, "github_v2", null) != null ? [lookup(auth_settings_v2.value, "github_v2")] : []

        content {
          client_id                  = github_v2.value.client_id
          client_secret_setting_name = github_v2.value.client_secret_setting_name
          login_scopes               = github_v2.value.login_scopes
        }
      }

      dynamic "google_v2" {
        for_each = lookup(auth_settings_v2.value, "google_v2", null) != null ? [lookup(auth_settings_v2.value, "google_v2")] : []

        content {
          client_id                  = google_v2.value.client_id
          client_secret_setting_name = google_v2.value.client_secret_setting_name
          allowed_audiences          = google_v2.value.allowed_audiences
          login_scopes               = google_v2.value.login_scopes
        }
      }

      dynamic "microsoft_v2" {
        for_each = lookup(auth_settings_v2.value, "microsoft_v2", null) != null ? [lookup(auth_settings_v2.value, "microsoft_v2")] : []

        content {
          client_id                  = microsoft_v2.value.client_id
          client_secret_setting_name = microsoft_v2.value.client_secret_setting_name
          allowed_audiences          = microsoft_v2.value.allowed_audiences
          login_scopes               = microsoft_v2.value.login_scopes
        }
      }

      dynamic "twitter_v2" {
        for_each = lookup(auth_settings_v2.value, "twitter_v2", null) != null ? [lookup(auth_settings_v2.value, "twitter_v2")] : []

        content {
          consumer_key                 = twitter_v2.value.consumer_key
          consumer_secret_setting_name = twitter_v2.value.consumer_secret_setting_name
        }
      }
    }
  }

  dynamic "storage_account" {
    for_each = lookup(
      each.value, "storage_accounts", {}
    )

    content {
      name = lookup(
        storage_account.value, "name", storage_account.key
      )

      type         = storage_account.value.type
      share_name   = storage_account.value.share_name
      access_key   = storage_account.value.access_key
      account_name = storage_account.value.account_name
      mount_path   = storage_account.value.mount_path
    }
  }

  dynamic "connection_string" {
    for_each = lookup(
      each.value, "connection_strings", {}
    )

    content {
      name = lookup(
        connection_string.value, "name", connection_string.key
      )

      value = connection_string.value
      type  = connection_string.type
    }
  }

  dynamic "backup" {
    for_each = lookup(each.value, "backup", null) != null ? [lookup(each.value, "backup")] : []

    content {
      name                = backup.value.name
      storage_account_url = backup.value.storage_account_url
      enabled             = backup.value.enabled

      schedule {
        frequency_interval       = backup.value.schedule.frequency_interval
        frequency_unit           = backup.value.schedule.frequency_unit
        keep_at_least_one_backup = backup.value.schedule.keep_at_least_one_backup
        retention_period_days    = backup.value.schedule.retention_period_days
        start_time               = backup.value.schedule.start_time
        last_execution_time      = backup.value.schedule.last_execution_time
      }
    }
  }

  site_config {
    always_on                              = each.value.site_config.always_on
    ftps_state                             = each.value.site_config.ftps_state
    worker_count                           = each.value.site_config.worker_count
    http2_enabled                          = each.value.site_config.http2_enabled
    app_scale_limit                        = each.value.site_config.app_scale_limit
    app_command_line                       = each.value.site_config.app_command_line
    remote_debugging_version               = each.value.site_config.remote_debugging_version
    pre_warmed_instance_count              = each.value.site_config.pre_warmed_instance_count
    runtime_scale_monitoring_enabled       = each.value.site_config.runtime_scale_monitoring_enabled
    scm_use_main_ip_restriction            = each.value.site_config.scm_use_main_ip_restriction
    health_check_eviction_time_in_min      = each.value.site_config.health_check_eviction_time_in_min
    application_insights_connection_string = each.value.site_config.application_insights_connection_string
    minimum_tls_version                    = each.value.site_config.minimum_tls_version
    api_management_api_id                  = each.value.site_config.api_management_api_id
    managed_pipeline_mode                  = each.value.site_config.managed_pipeline_mode
    vnet_route_all_enabled                 = each.value.site_config.vnet_route_all_enabled
    scm_minimum_tls_version                = each.value.site_config.scm_minimum_tls_version
    application_insights_key               = each.value.site_config.application_insights_key
    elastic_instance_minimum               = each.value.site_config.elastic_instance_minimum
    remote_debugging_enabled               = each.value.site_config.remote_debugging_enabled
    default_documents                      = each.value.site_config.default_documents
    health_check_path                      = each.value.site_config.health_check_path
    use_32_bit_worker                      = each.value.site_config.use_32_bit_worker
    api_definition_url                     = each.value.site_config.api_definition_url
    auto_swap_slot_name                    = each.value.site_config.auto_swap_slot_name
    websockets_enabled                     = each.value.site_config.websockets_enabled
    load_balancing_mode                    = each.value.site_config.load_balancing_mode
    ip_restriction_default_action          = each.value.site_config.ip_restriction_default_action
    scm_ip_restriction_default_action      = each.value.site_config.scm_ip_restriction_default_action

    dynamic "ip_restriction" {
      for_each = try(
        each.value.site_config.ip_restrictions, {}
      )

      content {
        action                    = ip_restriction.value.action
        ip_address                = ip_restriction.value.ip_address
        name                      = ip_restriction.value.name
        priority                  = ip_restriction.value.priority
        service_tag               = ip_restriction.value.service_tag
        virtual_network_subnet_id = ip_restriction.value.virtual_network_subnet_id
        description               = ip_restriction.value.description
        headers                   = ip_restriction.value.headers
      }
    }

    dynamic "scm_ip_restriction" {
      for_each = try(var.instance.site_config.scm_ip_restriction, {})

      content {
        action                    = scm_ip_restriction.value.action
        ip_address                = scm_ip_restriction.value.ip_address
        name                      = scm_ip_restriction.value.name
        priority                  = scm_ip_restriction.value.priority
        service_tag               = scm_ip_restriction.value.service_tag
        virtual_network_subnet_id = scm_ip_restriction.value.virtual_network_subnet_id
        headers                   = scm_ip_restriction.value.headers
        description               = scm_ip_restriction.value.description
      }
    }

    dynamic "application_stack" {
      for_each = lookup(each.value.site_config, "application_stack", null) != null ? [lookup(each.value.site_config, "application_stack")] : []

      content {
        dotnet_version              = application_stack.value.dotnet_version
        use_dotnet_isolated_runtime = application_stack.value.use_dotnet_isolated_runtime
        java_version                = application_stack.value.java_version
        node_version                = application_stack.value.node_version
        powershell_core_version     = application_stack.value.powershell_core_version
        use_custom_runtime          = application_stack.value.use_custom_runtime
      }
    }

    dynamic "cors" {
      for_each = lookup(each.value.site_config, "cors", null) != null ? [lookup(each.value.site_config, "cors")] : []

      content {
        allowed_origins     = cors.value.allowed_origins
        support_credentials = cors.value.support_credentials
      }
    }

    dynamic "app_service_logs" {
      for_each = lookup(each.value.site_config, "app_service_logs", null) != null ? [lookup(each.value.site_config, "app_service_logs")] : []

      content {
        disk_quota_mb         = app_service_logs.value.disk_quota_mb
        retention_period_days = app_service_logs.value.retention_period_days
      }
    }
  }
  lifecycle {
    ignore_changes = [
      auth_settings,
      app_settings["WEBSITE_RUN_FROM_PACKAGE"],
      app_settings["WEBSITE_ENABLE_SYNC_UPDATE_SITE"],
      tags["hidden-link: /app-insights-instrumentation-key"],
      tags["hidden-link: /app-insights-resource-id"],
      tags["hidden-link: /app-insights-conn-string"],
    ]
  }
}
