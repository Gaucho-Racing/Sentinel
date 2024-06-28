import { User, initUser } from "@/models/user";

export const currentUser: User = initUser;

export const SENTINEL_API_URL =
  import.meta.env.VITE_SENTINEL_API_URL ??
  "https://sentinel.gauchoracing.com/api";

export const DISCORD_CLIENT_ID = "1204930904913481840";
export const DISCORD_OAUTH_BASE_URL =
  "https://discord.com/api/oauth2/authorize";
export const DISCORD_SERVER_INVITE_URL = "https://discord.gg/tvYFre2m4F";
