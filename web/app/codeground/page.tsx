"use client";

import NewCodeground from "@/components/new-codeground";
import { Card } from "@/components/ui/card";
import api from "@/lib/api";
import { FolderGit2 } from "lucide-react";
import Link from "next/link";
import toast from "react-hot-toast";
import { useRouter } from "next/navigation";
import { codeground } from "@/types/codeground";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";

const fetchCodegrounds = async () => {
  const response = await api.get("/codeground");
  if (response.status !== 200) {
    toast.error("Failed to fetch codegrounds");
    throw new Error("Failed to fetch codegrounds");
  }
  return response.data.data;
};

export default function DashboardPage() {
  const router = useRouter();
  const queryClient = useQueryClient();

  const LogoutHandler = async () => {
    try {
      await api.post("/logout");
      toast.success("Logged out successfully");
      router.push("/");
    } catch (error) {
      toast.error(error as string);
    }
  };

  const {
    data: codegrounds = [],
    error,
    isLoading,
  } = useQuery({
    queryKey: ["codegrounds"],
    queryFn: fetchCodegrounds,
  });

  const addCodegroundToCache = (newCodeground: codeground) => {
    queryClient.setQueryData(
      ["codegrounds"],
      (oldData: codeground[] | undefined) => {
        return [...(oldData || []), newCodeground];
      }
    );
  };
  const handleNewCodeground = (newCodeground: codeground) => {
    addCodegroundToCache(newCodeground);
  };

  return (
    <div className="min-h-screen p-8 bg-background">
      <div className="mx-auto space-y-8 max-w-7xl">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold">My Codegrounds</h1>
          <div className="flex items-start gap-3">
            <NewCodeground onCodegroundCreated={handleNewCodeground} />
            <Button onClick={LogoutHandler}>Logout</Button>
          </div>
        </div>

        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {isLoading ? (
            <>
              <LoadingSkeleton />
              <LoadingSkeleton />
              <LoadingSkeleton />
            </>
          ) : codegrounds === null ? (
            <div className="text-2xl">No Codegrounds Found</div>
          ) : (
            codegrounds.map(
              codegrounds.map(
                (codeground: {
                  id: string;
                  codeground_type: string;
                  name: string;
                  updatedAt: string;
                  createdAt: string;
                  userId: number;
                }) => (
                  <Link
                    href={`/codeground/${codeground.id}`}
                    key={codeground.id}
                  >
                    <Card className="p-6 transition-colors cursor-pointer hover:border-primary">
                      <div className="flex items-start justify-between">
                        <div className="space-y-1">
                          <FolderGit2 className="w-8 h-8 mb-2 text-primary" />
                          <h2 className="font-semibold">{codeground.name}</h2>
                        </div>
                        <p className="text-xs text-muted-foreground">
                          {codeground.codeground_type}
                        </p>
                      </div>
                    </Card>
                  </Link>
                )
              )
            )
          )}
        </div>
      </div>
    </div>
  );
}

const LoadingSkeleton = () => (
  <Card className="p-6">
    <div className="flex items-start justify-between">
      <div className="space-y-3">
        <Skeleton className="w-8 h-8 rounded" />
        <Skeleton className="h-4 w-[150px]" />
        <Skeleton className="h-4 w-[100px]" />
      </div>
      <Skeleton className="h-4 w-[100px]" />
    </div>
  </Card>
);
