import axios from "axios";
import { apiBaseUrl } from "@/lib/config";

const api = axios.create({
  baseURL: apiBaseUrl,
  withCredentials: true,
  headers: {
    "Content-Type": "application/json",
  },
});

export default api;
