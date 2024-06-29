import { SENTINEL_API_URL, currentUser } from "@/consts/config";
import { initUser, setUser } from "@/models/user";
import axios from "axios";

export const checkCredentials = async () => {
  if (
    localStorage.getItem("id") == null ||
    localStorage.getItem("token") == null
  ) {
    return 1;
  } else if (currentUser.id == "") {
    try {
      const userId = localStorage.getItem("id");
      const response = await axios.get(`${SENTINEL_API_URL}/users/${userId}`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem("token")}`,
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
  localStorage.removeItem("id");
  localStorage.removeItem("token");
  setUser(currentUser, initUser);
};
