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
  setUser(currentUser, initUser);
};
