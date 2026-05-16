// Application API shape — mirror of core's model.Application JSON.
// owner_id is the entity_id of the creator (USER or SERVICE_ACCOUNT entity).
export type Application = {
  id: string
  owner_id: string
  name: string
  description: string
  client_id: string
  icon_url: string
  launch_url: string
  redirect_uris: string[]
  updated_at: string
  created_at: string
}
