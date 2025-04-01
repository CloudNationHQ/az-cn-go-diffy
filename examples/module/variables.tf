variable "instance" {
  description = "Contains all function app configuration"
  type = object({
    name                                           = string
    type                                           = string # "linux" or "windows"
    resource_group                                 = optional(string, null)
    location                                       = optional(string, null)
    service_plan_id                                = string
    storage_account_name                           = string
    storage_account_access_key                     = string
    https_only                                     = optional(bool, true)
    zip_deploy_file                                = optional(string, null)
    enabled                                        = optional(bool, true)
    builtin_logging_enabled                        = optional(bool, true)
    client_certificate_mode                        = optional(string, null)
    daily_memory_time_quota                        = optional(number, null)
    virtual_network_subnet_id                      = optional(string, null)
    client_certificate_enabled                     = optional(bool, false)
    functions_extension_version                    = optional(string, null)
    storage_key_vault_secret_id                    = optional(string, null)
    content_share_force_disabled                   = optional(bool, false)
    public_network_access_enabled                  = optional(bool, true)
    storage_uses_managed_identity                  = optional(bool, null)
    vnet_image_pull_enabled                        = optional(bool, false)
    key_vault_reference_identity_id                = optional(string, null)
    client_certificate_exclusion_paths             = optional(string, null)
    ftp_publish_basic_authentication_enabled       = optional(bool, true)
    webdeploy_publish_basic_authentication_enabled = optional(bool, true)
    virtual_network_backup_restore_enabled         = optional(bool, false)
    app_settings                                   = optional(map(string), null)
    tags                                           = optional(map(string))
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), null)
    }), null)
    auth_settings_v2 = optional(object({
      auth_enabled                            = optional(bool, false)
      runtime_version                         = optional(string, "~1")
      config_file_path                        = optional(string, null)
      require_authentication                  = optional(bool, null)
      unauthenticated_action                  = optional(string, "RedirectToLoginPage")
      default_provider                        = optional(string, null)
      excluded_paths                          = optional(list(string), null)
      require_https                           = optional(bool, true)
      http_route_api_prefix                   = optional(string, "/.auth")
      forward_proxy_convention                = optional(string, "NoProxy")
      forward_proxy_custom_host_header_name   = optional(string, null)
      forward_proxy_custom_scheme_header_name = optional(string, null)
      login = object({
        token_store_enabled               = optional(bool, false)
        token_refresh_extension_time      = optional(number, null)
        token_store_path                  = optional(string, null)
        token_store_sas_setting_name      = optional(string, null)
        preserve_url_fragments_for_logins = optional(bool, null)
        allowed_external_redirect_urls    = optional(list(string), null)
        cookie_expiration_convention      = optional(string, null)
        cookie_expiration_time            = optional(string, null)
        validate_nonce                    = optional(bool, null)
        nonce_expiration_time             = optional(string, null)
        logout_endpoint                   = optional(string, null)
      })
      apple_v2 = optional(object({
        client_id                  = string
        client_secret_setting_name = string
        login_scopes               = optional(list(string), null)
      }), null)
      active_directory_v2 = optional(object({
        client_id                            = string
        tenant_auth_endpoint                 = string
        client_secret_setting_name           = optional(string, null)
        client_secret_certificate_thumbprint = optional(string, null)
        jwt_allowed_groups                   = optional(list(string), null)
        jwt_allowed_client_applications      = optional(list(string), null)
        www_authentication_disabled          = optional(bool, null)
        allowed_audiences                    = optional(list(string), null)
        allowed_groups                       = optional(list(string), null)
        allowed_identities                   = optional(list(string), null)
        login_parameters                     = optional(map(string), null)
        allowed_applications                 = optional(list(string), null)
      }), null)
      azure_static_web_app_v2 = optional(object({
        client_id = string
      }), null)
      custom_oidc_v2 = optional(map(object({
        name                          = string
        client_id                     = string
        openid_configuration_endpoint = string
        name_claim_type               = optional(string, null)
        scopes                        = optional(list(string), null)
        client_credential_method      = optional(string, null)
        client_secret_setting_name    = optional(string, null)
        authorisation_endpoint        = optional(string, null)
        token_endpoint                = optional(string, null)
        issuer_endpoint               = optional(string, null)
        certification_uri             = optional(string, null)
      })), {})
      facebook_v2 = optional(object({
        app_id                  = string
        app_secret_setting_name = string
        graph_api_version       = optional(string, "v15.0")
        login_scopes            = optional(list(string), null)
      }), null)
      github_v2 = optional(object({
        client_id                  = string
        client_secret_setting_name = string
        login_scopes               = optional(list(string), null)
      }), null)
      google_v2 = optional(object({
        client_id                  = string
        client_secret_setting_name = string
        allowed_audiences          = optional(list(string), null)
        login_scopes               = optional(list(string), null)
      }), null)
      microsoft_v2 = optional(object({
        client_id                  = string
        client_secret_setting_name = string
        allowed_audiences          = optional(list(string), null)
        login_scopes               = optional(list(string), null)
      }), null)
      twitter_v2 = optional(object({
        consumer_key                 = string
        consumer_secret_setting_name = string
      }), null)
    }), null)
    site_config = object({
      always_on                                     = optional(bool, false)
      ftps_state                                    = optional(string, "AllAllowed")
      worker_count                                  = optional(number, null)
      http2_enabled                                 = optional(bool, true)
      app_scale_limit                               = optional(number, null)
      app_command_line                              = optional(string, null)
      remote_debugging_version                      = optional(string, null)
      pre_warmed_instance_count                     = optional(number, null)
      runtime_scale_monitoring_enabled              = optional(bool, false)
      scm_use_main_ip_restriction                   = optional(bool, false)
      health_check_eviction_time_in_min             = optional(number, null)
      application_insights_connection_string        = optional(string, null)
      container_registry_use_managed_identity       = optional(bool, false)
      container_registry_managed_identity_client_id = optional(string, null)
      minimum_tls_version                           = optional(string, "1.2")
      api_management_api_id                         = optional(string, null)
      managed_pipeline_mode                         = optional(string, null)
      vnet_route_all_enabled                        = optional(bool, false)
      scm_minimum_tls_version                       = optional(string, "1.2")
      application_insights_key                      = optional(string, null)
      elastic_instance_minimum                      = optional(number, null)
      remote_debugging_enabled                      = optional(bool, false)
      default_documents                             = optional(list(string), null)
      health_check_path                             = optional(string, null)
      use_32_bit_worker                             = optional(bool, false)
      api_definition_url                            = optional(string, null)
      websockets_enabled                            = optional(bool, false)
      load_balancing_mode                           = optional(string, null)
      ip_restriction_default_action                 = optional(string, "Allow")
      scm_ip_restriction_default_action             = optional(string, "Allow")
      ip_restrictions = optional(map(object({
        action                    = optional(string, "Allow")
        ip_address                = optional(string, null)
        name                      = optional(string, null)
        priority                  = optional(number, 65000)
        service_tag               = optional(string, null)
        virtual_network_subnet_id = optional(string, null)
        description               = optional(string, null)
        headers                   = optional(list(string), [])
      })), {})
      scm_ip_restrictions = optional(map(object({
        action                    = optional(string, "Allow")
        ip_address                = optional(string, null)
        name                      = optional(string, null)
        priority                  = optional(number, 65000)
        service_tag               = optional(string, null)
        virtual_network_subnet_id = optional(string, null)
        description               = optional(string, null)
        headers                   = optional(list(string), [])
      })), {})
      application_stack = optional(object({
        dotnet_version              = optional(string, null)
        use_dotnet_isolated_runtime = optional(bool, null)
        java_version                = optional(string, null)
        node_version                = optional(string, null)
        python_version              = optional(string, null)
        powershell_core_version     = optional(string, null)
        use_custom_runtime          = optional(bool, null)
        docker = optional(object({
          image_name        = string
          image_tag         = string
          registry_url      = string
          registry_username = string
          registry_password = string
        }), null)
      }), null)
      cors = optional(object({
        allowed_origins     = optional(list(string), [])
        support_credentials = optional(bool, false)
      }), null)
      app_service_logs = optional(object({
        disk_quota_mb         = optional(number, null)
        retention_period_days = optional(number, null)
      }), null)
    })
    storage_accounts = optional(map(object({
      name         = optional(string, null)
      type         = string
      share_name   = string
      access_key   = string
      account_name = string
      mount_path   = optional(string, null)
    })), {})
    sticky_settings = optional(object({
      app_setting_names       = optional(list(string), [])
      connection_string_names = optional(list(string), [])
    }), null)
    backup = optional(object({
      name                = string
      enabled             = optional(bool, true)
      storage_account_url = string
      schedule = object({
        frequency_unit           = string
        frequency_interval       = number
        retention_period_days    = optional(number, null)
        start_time               = optional(string, null)
        keep_at_least_one_backup = optional(bool, false)
      })
    }), null)
    connection_string = optional(map(object({
      name  = string
      type  = string
      value = string
    })), {})
    slots = optional(map(object({
      name                                   = string
      service_plan_id                        = optional(string, null)
      virtual_network_backup_restore_enabled = optional(bool, false)
      app_settings                           = optional(map(string), null)
      identity = optional(object({
        type         = string
        identity_ids = optional(list(string), null)
      }), null)
      site_config = object({
        always_on                                     = optional(bool, false)
        ftps_state                                    = optional(string, "Disabled")
        worker_count                                  = optional(number, null)
        http2_enabled                                 = optional(bool, false)
        app_scale_limit                               = optional(number, null)
        app_command_line                              = optional(string, null)
        remote_debugging_version                      = optional(string, null)
        pre_warmed_instance_count                     = optional(number, null)
        runtime_scale_monitoring_enabled              = optional(bool, false)
        scm_use_main_ip_restriction                   = optional(bool, false)
        health_check_eviction_time_in_min             = optional(number, null)
        application_insights_connection_string        = optional(string, null)
        container_registry_use_managed_identity       = optional(bool, false)
        container_registry_managed_identity_client_id = optional(string, null)
        minimum_tls_version                           = optional(string, "1.2")
        api_management_api_id                         = optional(string, null)
        managed_pipeline_mode                         = optional(string, null)
        vnet_route_all_enabled                        = optional(bool, false)
        scm_minimum_tls_version                       = optional(string, "1.2")
        application_insights_key                      = optional(string, null)
        elastic_instance_minimum                      = optional(number, null)
        remote_debugging_enabled                      = optional(bool, false)
        default_documents                             = optional(list(string), null)
        health_check_path                             = optional(string, null)
        use_32_bit_worker                             = optional(bool, false)
        api_definition_url                            = optional(string, null)
        auto_swap_slot_name                           = optional(string, null)
        websockets_enabled                            = optional(bool, false)
        load_balancing_mode                           = optional(string, null)
        scm_ip_restriction_default_action             = optional(string, "Allow")
        ip_restriction_default_action                 = optional(string, "Allow")
        application_stack = optional(object({
          dotnet_version              = optional(string, null)
          use_dotnet_isolated_runtime = optional(bool, null)
          java_version                = optional(string, null)
          node_version                = optional(string, null)
          python_version              = optional(string, null)
          powershell_core_version     = optional(string, null)
          use_custom_runtime          = optional(bool, null)
          docker = optional(object({
            image_name        = string
            image_tag         = string
            registry_url      = string
            registry_username = string
            registry_password = string
          }), null)
        }), null)
        cors = optional(object({
          allowed_origins     = optional(list(string), [])
          support_credentials = optional(bool, false)
        }), null)
        app_service_logs = optional(object({
          disk_quota_mb         = optional(number, null)
          retention_period_days = optional(number, null)
        }), null)
      })
      auth_settings_v2 = optional(object({
        auth_enabled                            = optional(bool, false)
        runtime_version                         = optional(string, "~1")
        config_file_path                        = optional(string, null)
        require_authentication                  = optional(bool, null)
        unauthenticated_action                  = optional(string, "RedirectToLoginPage")
        default_provider                        = optional(string, null)
        excluded_paths                          = optional(list(string), null)
        require_https                           = optional(bool, true)
        http_route_api_prefix                   = optional(string, "/.auth")
        forward_proxy_convention                = optional(string, "NoProxy")
        forward_proxy_custom_host_header_name   = optional(string, null)
        forward_proxy_custom_scheme_header_name = optional(string, null)
        login = object({
          token_store_enabled               = optional(bool, false)
          token_refresh_extension_time      = optional(number, null)
          token_store_path                  = optional(string, null)
          token_store_sas_setting_name      = optional(string, null)
          preserve_url_fragments_for_logins = optional(bool, null)
          allowed_external_redirect_urls    = optional(list(string), null)
          cookie_expiration_convention      = optional(string, null)
          cookie_expiration_time            = optional(string, null)
          validate_nonce                    = optional(bool, null)
          nonce_expiration_time             = optional(string, null)
          logout_endpoint                   = optional(string, null)
        })
        apple_v2 = optional(object({
          client_id                  = string
          client_secret_setting_name = string
          login_scopes               = optional(list(string), null)
        }), null)
        active_directory_v2 = optional(object({
          client_id                            = string
          tenant_auth_endpoint                 = string
          client_secret_setting_name           = optional(string, null)
          client_secret_certificate_thumbprint = optional(string, null)
          jwt_allowed_groups                   = optional(list(string), null)
          jwt_allowed_client_applications      = optional(list(string), null)
          www_authentication_disabled          = optional(bool, null)
          allowed_audiences                    = optional(list(string), null)
          allowed_groups                       = optional(list(string), null)
          allowed_identities                   = optional(list(string), null)
          login_parameters                     = optional(map(string), null)
          allowed_applications                 = optional(list(string), null)
        }), null)
        azure_static_web_app_v2 = optional(object({
          client_id = string
        }), null)
        custom_oidc_v2 = optional(map(object({
          name                          = string
          client_id                     = string
          openid_configuration_endpoint = string
          name_claim_type               = optional(string, null)
          scopes                        = optional(list(string), null)
          client_credential_method      = optional(string, null)
          client_secret_setting_name    = optional(string, null)
          authorisation_endpoint        = optional(string, null)
          token_endpoint                = optional(string, null)
          issuer_endpoint               = optional(string, null)
          certification_uri             = optional(string, null)
        })), {})
        facebook_v2 = optional(object({
          app_id                  = string
          app_secret_setting_name = string
          graph_api_version       = optional(string, "v15.0")
          login_scopes            = optional(list(string), null)
        }), null)
        github_v2 = optional(object({
          client_id                  = string
          client_secret_setting_name = string
          login_scopes               = optional(list(string), null)
        }), null)
        google_v2 = optional(object({
          client_id                  = string
          client_secret_setting_name = string
          allowed_audiences          = optional(list(string), null)
          login_scopes               = optional(list(string), null)
        }), null)
        microsoft_v2 = optional(object({
          client_id                  = string
          client_secret_setting_name = string
          allowed_audiences          = optional(list(string), null)
          login_scopes               = optional(list(string), null)
        }), null)
        twitter_v2 = optional(object({
          consumer_key                 = string
          consumer_secret_setting_name = string
        }), null)
      }), null)
      storage_accounts = optional(map(object({
        name         = optional(string, null)
        type         = string
        share_name   = string
        access_key   = string
        account_name = string
        mount_path   = optional(string, null)
      })), {})
      connection_strings = optional(map(object({
        name  = optional(string, null)
        value = string
        type  = string
      })), {})
      backup = optional(object({
        name                = string
        storage_account_url = string
        enabled             = optional(bool, true)
        schedule = object({
          frequency_interval       = number
          frequency_unit           = string
          keep_at_least_one_backup = optional(bool, false)
          retention_period_days    = optional(number, 30)
          start_time               = optional(string, null)
          last_execution_time      = optional(string, null)
        })
      }), null)
    })), {})
  })
  validation {
    condition     = var.instance.location != null || var.location != null
    error_message = "location must be provided either in the config object or as a separate variable."
  }

  validation {
    condition     = var.instance.resource_group != null || var.resource_group != null
    error_message = "resource group name must be provided either in the config object or as a separate variable."
  }

  validation {
    condition     = contains(["windows", "linux"], lookup(var.instance, "type", ""))
    error_message = "The vm type must be either 'windows' or 'linux'."
  }
}

variable "resource_group" {
  description = "Default resource group name"
  type        = string
  default     = null
}

variable "location" {
  description = "Default location"
  type        = string
  default     = null
}

variable "tags" {
  description = "Default tags"
  type        = map(string)
  default     = {}
}
