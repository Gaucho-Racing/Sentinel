import { User, initUser } from "@/models/user";

export const currentUser: User = initUser;

export const SENTINEL_API_URL =
  import.meta.env.VITE_SENTINEL_API_URL ??
  "https://sentinel.gauchoracing.com/api";
