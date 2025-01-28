"use server";
import axios from "axios";
import { LoginFormValues } from "@/types/auth";

export async function loginUser(data: LoginFormValues) {
  try {
    const response = await axios.post("http://localhost:8080/api/login", data);
    return { success: true, data: response.data };
  } catch (error) {
    return { success: false, error: "Login failed" };
  }
}
