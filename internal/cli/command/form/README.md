# Example

## Polulate config file

```sh
cat > form-config.yaml <<EOF
form:
  - description:
      Credentials are required to support price comparisons of infrastructures
      on different cloud providers
    fields:
      - controlType: checkbox
        default: false
        key: enable_google
        label: enable
        placeholder: Google
        value: false
      - controlGroupType: google
        key: google
        showIf:
          properties:
            enable_google:
              const: true
      - controlType: checkbox
        default: false
        key: enable_amazon
        label: enable
        placeholder: Amazon
        value: false
      - controlGroupType: amazon
        key: amazon
        showIf:
          properties:
            enable_amazon:
              const: true
      - controlType: checkbox
        default: false
        key: enable_alibaba
        label: enable
        placeholder: Alibaba
        value: false
      - controlGroupType: alibaba
        key: alibaba
        showIf:
          properties:
            enable_alibaba:
              const: true
      - controlType: checkbox
        default: false
        key: enable_azure
        label: enable
        placeholder: Azure
      - controlGroupType: azure
        key: azure
        showIf:
          properties:
            enable_azure:
              const: true
      - controlType: checkbox
        default: false
        key: enable_oracle
        label: enable
        placeholder: Oracle
      - controlGroupType: oracle
        key: oracle
        showIf:
          properties:
            enable_oracle:
              const: true
    name: Provider credentials
  - fields:
      - controlType: text
        description: Fully qualified domain name
        key: cloudinfo_fqdn
        label: FQDN
        pattern: ^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.){2,}([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9]){2,}$
        placeholder: www.example.com
        required: true
    link: https://github.com/banzaicloud/cloudinfo
    name: Cloud Info
  - fields:
      - controlType: text
        description: Fully qualified domain name
        key: recommender_fqdn
        label: FQDN
        pattern: ^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.){2,}([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9]){2,}$
        placeholder: www.example.com
        required: true
    link: https://github.com/banzaicloud/telescopes
    name: Recommender
templates:
  values.yaml: |
    app:
      basePath: /recommender
      cloudInfoAddress: http://{{ .recommender_fqdn }}/cloudinfo/api/v1/
    ingress:
      enabled: true
      hosts:
      - {{ .recommender_fqdn }}/recommender
    {{- if .enable_google }}
    google:
      key: |
        {{ .google.json_key | nindent 4 }}
    {{- end }}
EOF
```

## Open form

```sh
# banzai form open FORM_CONFIG [--port 0] [--browser]
$ banzai form open ./form-config.yaml --browser
```

## Render template

```sh
# banzai form template FORM_CONFIG [-n TEMPLATE_NAME] [--force]
$ banzai form template ./form-config.yaml
```

## Migrate config values

```sh
# banzai form migrate SOURCE_FORM_CONFIG TARGET_FORM_CONFIG
$ cp ./form-config.yaml ./form-config-new.yaml
$ banzai form migrate ./form-config.yaml ./form-config-new.yaml
```
