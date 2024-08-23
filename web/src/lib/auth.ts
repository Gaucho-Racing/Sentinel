import { SENTINEL_API_URL, currentUser } from "@/consts/config";
import { initUser, setUser } from "@/models/user";
import axios from "axios";

export const checkCredentials = async () => {
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
        setUser(currentUser, response.data);
        return 0;
      }
    } catch (error) {
      logout();
      return 1;
    }
  } else {
    return 0;
  }
  return 1;
};

export const logout = () => {
  localStorage.removeItem("sentinel_access_token");
  // Remove all cookies that start with sentinel_
  document.cookie.split(";").forEach((cookie) => {
    const trimmedCookie = cookie.trim();
    if (trimmedCookie.startsWith("sentinel_")) {
      const cookieName = trimmedCookie.split("=")[0];
      document.cookie = `${cookieName}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/; domain=.gauchoracing.com; secure; samesite=lax`;
    }
  });
  setUser(currentUser, initUser);
};

export const saveAccessToken = (accessToken: string) => {
  localStorage.setItem("sentinel_access_token", accessToken);
  document.cookie = `sentinel_access_token=${accessToken}; domain=.gauchoracing.com; path=/; secure; samesite=lax`;
};
