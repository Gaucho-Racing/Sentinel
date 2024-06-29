import { User, initUser } from "@/models/user";

export let currentUser: User = Object.assign({}, initUser);

export const SENTINEL_API_URL =
  import.meta.env.VITE_SENTINEL_API_URL ??
  "https://sentinel.gauchoracing.com/api";

export const DISCORD_CLIENT_ID = "1204930904913481840";
export const DISCORD_OAUTH_BASE_URL =
  "https://discord.com/api/oauth2/authorize";
export const DISCORD_SERVER_INVITE_URL = "https://discord.gg/tvYFre2m4F";

export const SHARED_DRIVE_URL =
  "https://drive.google.com/drive/u/0/folders/0ADMP93ZBlor_Uk9PVA";
export const GITHUB_ORG_URL = "https://github.com/Gaucho-Racing";

export const SOCIAL_LINKS = {
  github: "https://github.com/gaucho-racing/sentinel",
  instagram: "https://instagram.com/gauchoracingucsb",
  twitter: "https://twitter.com/gaucho_racing",
  linkedin:
    "https://www.linkedin.com/company/gaucho-racing-at-uc-santa-barbara",
};
