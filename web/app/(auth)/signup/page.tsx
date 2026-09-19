"use client";

import { Button } from "@/components/ui/button";
import Link from "next/link";
import api from "@/lib/api";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { Form } from "@/components/ui/form";
import { FormTextField } from "@/components/auth/FormTextField";
import { SignupFormValues, signupSchema } from "@/types/auth";
import toast from "react-hot-toast";

export default function SignupPage() {
  const [error, setError] = useState<string | undefined>("");
  const router = useRouter();
  const form = useForm<SignupFormValues>({
    resolver: zodResolver(signupSchema),
    defaultValues: {
      email: "",
      password: "",
      confirmPassword: "",
    },
  });

  const onSubmit = async (data: SignupFormValues) => {
    try {
      const result = await api.post("/signup", {
        email: data.email,
        password: data.password,
      });
      toast.success("Signup successful");
      router.push("/codeground");
    } catch (error) {
      setError("An unexpected error occurred");
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <div className="w-full max-w-md p-8 space-y-6 bg-card rounded-lg border">
        <div className="space-y-2 text-center">
          <h1 className="text-3xl font-bold">Create an account</h1>
          <p className="text-muted-foreground">
            Enter your details to get started
          </p>
        </div>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormTextField
              control={form.control}
              name="email"
              label="Email"
              type="email"
              placeholder="Enter your email"
            />

            <FormTextField
              control={form.control}
              name="password"
              label="Password"
              type="password"
              placeholder="Create a password"
            />

            <FormTextField
              control={form.control}
              name="confirmPassword"
              label="Confirm Password"
              type="password"
              placeholder="Confirm your password"
            />

            <Button
              type="submit"
              className="w-full"
              disabled={form.formState.isSubmitting}
            >
              {form.formState.isSubmitting
                ? "Creating account..."
                : "Create account"}
            </Button>

            {error && (
              <p className="text-sm text-red-500 text-center">{error}</p>
            )}

            <p className="text-sm text-center text-muted-foreground">
              Already have an account?{" "}
              <Link href="/login" className="text-primary hover:underline">
                Sign in
              </Link>
            </p>
          </form>
        </Form>
      </div>
    </div>
  );
}
