"use client";

import NewCodeground from "@/components/new-codeground";
import { Card } from "@/components/ui/card";
import api from "@/lib/api";
import { FolderGit2, Trash2 } from "lucide-react";
import Link from "next/link";
import toast from "react-hot-toast";
import { useRouter } from "next/navigation";
import { codeground } from "@/types/codeground";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";

type CodegroundItem = {
  id: string;
  codeground_type: string;
  name: string;
  updatedAt: string;
  createdAt: string;
  userId: number;
};

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
      },
    );
  };
  const handleNewCodeground = (newCodeground: codeground) => {
    addCodegroundToCache(newCodeground);
  };

  const removeCodegroundFromCache = (id: string) => {
    queryClient.setQueryData(
      ["codegrounds"],
      (oldData: CodegroundItem[] | undefined) =>
        (oldData || []).filter((cg) => cg.id !== id),
    );
  };

  const handleDelete = async (id: string) => {
    try {
      await api.delete(`/codeground/${id}`);
      removeCodegroundFromCache(id);
      toast.success("Codeground deleted");
    } catch {
      toast.error("Failed to delete codeground");
    }
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
          ) : error ? (
            <div className="text-2xl text-destructive">
              Failed to load codegrounds. Please try again.
            </div>
          ) : !codegrounds || codegrounds.length === 0 ? (
            <div className="text-2xl">No Codegrounds Found</div>
          ) : (
            codegrounds.map((codeground: CodegroundItem) => (
              <CodegroundCard
                key={codeground.id}
                codeground={codeground}
                onDelete={handleDelete}
              />
            ))
          )}
        </div>
      </div>
    </div>
  );
}

const CodegroundCard = ({
  codeground,
  onDelete,
}: {
  codeground: CodegroundItem;
  onDelete: (id: string) => void;
}) => (
  <Card className="relative p-6 transition-colors hover:border-primary">
    <Link
      href={{
        pathname: `/codeground/${codeground.id}`,
        query: { data: JSON.stringify(codeground) },
      }}
      className="block"
    >
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <FolderGit2 className="w-8 h-8 mb-2 text-primary" />
          <h2 className="font-semibold">{codeground.name}</h2>
        </div>
        <p className="text-xs text-muted-foreground">
          {codeground.codeground_type}
        </p>
      </div>
      {codeground.updatedAt && (
        <p className="mt-4 text-xs text-muted-foreground">
          Updated {new Date(codeground.updatedAt).toLocaleDateString()}
        </p>
      )}
    </Link>

    <div className="absolute bottom-3 right-3">
      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-muted-foreground hover:text-destructive"
          >
            <Trash2 className="w-4 h-4" />
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete this codeground?</AlertDialogTitle>
            <AlertDialogDescription>
              This permanently deletes &quot;{codeground.name}&quot; and tears
              down its running environment. This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={() => onDelete(codeground.id)}>
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  </Card>
);

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
