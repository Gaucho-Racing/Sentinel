import { SENTINEL_API_URL } from "@/consts/config";
import { initUser } from "@/models/user";
import { getUser, setUser } from "@/lib/store";
import axios from "axios";

export const checkCredentials = async (): Promise<number> => {
  const currentUser = getUser();
  if (localStorage.getItem("sentinel_access_token") == null) {
    return 1;
  } else if (currentUser.id == "") {
    try {
      const response = await axios.get(`${SENTINEL_API_URL}/users/@me`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem("sentinel_access_token")}`,
        },
      });
      if (response.status == 200) {
        setUser(response.data);
        return 0;
      }
    } catch (error) {
      if ((await refreshAccessToken()) == 0) {
        return checkCredentials();
      }
      return 1;
    }
  }
  return 0;
};

const refreshAccessToken = async (): Promise<number> => {
  const refreshToken = localStorage.getItem("sentinel_refresh_token");
  if (refreshToken == null) {
    return 1;
  }
  try {
    const response = await axios.post(`${SENTINEL_API_URL}/oauth/token`, {
      grant_type: "refresh_token",
      refresh_token: refreshToken,
    });
    if (response.status == 200) {
      saveAccessToken(response.data.access_token);
      saveRefreshToken(response.data.refresh_token);
      return 0;
    }
  } catch (error) {
    return 1;
  }
  return 1;
};

export const logout = () => {
  localStorage.removeItem("sentinel_access_token");
  localStorage.removeItem("sentinel_refresh_token");
  // Remove all cookies that start with sentinel_
  document.cookie.split(";").forEach((cookie) => {
    const trimmedCookie = cookie.trim();
    if (trimmedCookie.startsWith("sentinel_")) {
      const cookieName = trimmedCookie.split("=")[0];
      document.cookie = `${cookieName}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/; domain=.gauchoracing.com; secure; samesite=lax`;
    }
  });
  setUser(initUser);
};

export const saveAccessToken = (accessToken: string) => {
  localStorage.setItem("sentinel_access_token", accessToken);
  document.cookie = `sentinel_access_token=${accessToken}; domain=.gauchoracing.com; path=/; secure; samesite=lax`;
};

export const saveRefreshToken = (refreshToken: string) => {
  localStorage.setItem("sentinel_refresh_token", refreshToken);
  document.cookie = `sentinel_refresh_token=${refreshToken}; domain=.gauchoracing.com; path=/; secure; samesite=lax`;
};
