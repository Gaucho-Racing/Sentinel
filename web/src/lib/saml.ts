export type SAMLProfile = "GENERIC" | "AWS_IDENTITY_CENTER"

export type SAMLNameIDSource = "EMAIL" | "USERNAME" | "ENTITY_ID"

export type SAMLAttributeSource =
  | "EMAIL"
  | "USERNAME"
  | "FIRST_NAME"
  | "LAST_NAME"
  | "DISPLAY_NAME"
  | "ENTITY_ID"
  | "GROUP_NAMES"
  | "GROUP_IDS"
  | "CONSTANT"

export type SAMLAttributeMapping = {
  name: string
  friendly_name: string
  name_format: string
  source: SAMLAttributeSource
  constant: string
  omit_if_empty: boolean
}

export type SAMLConfiguration = {
  application_id: string
  profile: SAMLProfile
  entity_id: string
  acs_url: string
  name_id_source: SAMLNameIDSource
  name_id_format: string
  attribute_mappings: SAMLAttributeMapping[]
  certificate_pem: string
  want_authn_requests_signed: boolean
  metadata_xml: string
  updated_at?: string
  created_at?: string
}

export type SAMLAssertionPreview = {
  name_id: string
  name_id_format: string
  attributes: Array<{
    name: string
    friendly_name: string
    name_format: string
    values: string[]
  }>
}

export const SAML_NAME_ID_FORMAT_EMAIL =
  "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
export const SAML_NAME_ID_FORMAT_PERSISTENT =
  "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent"
export const SAML_NAME_ID_FORMAT_UNSPECIFIED =
  "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified"
export const SAML_ATTRIBUTE_FORMAT_BASIC =
  "urn:oasis:names:tc:SAML:2.0:attrname-format:basic"
export const SAML_ATTRIBUTE_FORMAT_URI =
  "urn:oasis:names:tc:SAML:2.0:attrname-format:uri"
export const SAML_ATTRIBUTE_FORMAT_UNSPECIFIED =
  "urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified"

export const GENERIC_SAML_ATTRIBUTE_MAPPINGS: SAMLAttributeMapping[] = [
  {
    name: "email",
    friendly_name: "email",
    name_format: SAML_ATTRIBUTE_FORMAT_BASIC,
    source: "EMAIL",
    constant: "",
    omit_if_empty: true,
  },
  {
    name: "first_name",
    friendly_name: "first_name",
    name_format: SAML_ATTRIBUTE_FORMAT_BASIC,
    source: "FIRST_NAME",
    constant: "",
    omit_if_empty: true,
  },
  {
    name: "last_name",
    friendly_name: "last_name",
    name_format: SAML_ATTRIBUTE_FORMAT_BASIC,
    source: "LAST_NAME",
    constant: "",
    omit_if_empty: true,
  },
  {
    name: "groups",
    friendly_name: "groups",
    name_format: SAML_ATTRIBUTE_FORMAT_BASIC,
    source: "GROUP_NAMES",
    constant: "",
    omit_if_empty: true,
  },
  {
    name: "entity_id",
    friendly_name: "entity_id",
    name_format: SAML_ATTRIBUTE_FORMAT_BASIC,
    source: "ENTITY_ID",
    constant: "",
    omit_if_empty: false,
  },
]

export function newSAMLConfiguration(applicationID: string): SAMLConfiguration {
  return {
    application_id: applicationID,
    profile: "GENERIC",
    entity_id: "",
    acs_url: "",
    name_id_source: "EMAIL",
    name_id_format: SAML_NAME_ID_FORMAT_EMAIL,
    attribute_mappings: GENERIC_SAML_ATTRIBUTE_MAPPINGS.map((mapping) => ({ ...mapping })),
    certificate_pem: "",
    want_authn_requests_signed: false,
    metadata_xml: "",
  }
}
