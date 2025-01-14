project_id               = "1071548024491"
project_name             = "armored-witness"
signing_keyring_location = "global"
tf_state_location        = "europe-west2"

tls          = true
serve_domain = "api.transparency.dev"

lb_name = "transparency-dev-lb"

# TODO(mhutchinson): this is the old env and should be switched to the following
distributor_prod_host = "distributor-service-prod-oxxl2d5jeq-uc.a.run.app"
distributor_prod_port = 443

distributor_ci_host = "distributor-service-ci-oxxl2d5jeq-uc.a.run.app"
distributor_ci_port = 443

distributor_dev_host = "distributor-service-dev-oxxl2d5jeq-uc.a.run.app"
distributor_dev_port = 443
