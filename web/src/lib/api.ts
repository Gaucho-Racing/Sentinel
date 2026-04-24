import axios from "axios"

export const coreApi = axios.create({
  baseURL: import.meta.env.VITE_CORE_API_URL,
  withCredentials: false,
})

export const oauthApi = axios.create({
  baseURL: import.meta.env.VITE_OAUTH_API_URL,
  withCredentials: false,
})
