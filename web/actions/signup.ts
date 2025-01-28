"use server";
import { SignupFormValues } from "@/types/auth";
import axios from "axios";

export async function signupUser(data: SignupFormValues) {
  "use server";

  try {
    const response = await axios.post("http://localhost:8080/api/signup", data);
    return { success: true, data: response.data };
  } catch (error) {
    return { success: false, error: "Signup failed" };
  }
}
