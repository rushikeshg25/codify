"use client";

import NewCodeground from "@/components/new-codeground";
import { Card } from "@/components/ui/card";
import api from "@/lib/api";
import { FolderGit2, Trash2, Play, Square, Loader2 } from "lucide-react";
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

type CodegroundStatus =
  | "PROVISIONING"
  | "RUNNING"
  | "STOPPED"
  | "FAILED";

type CodegroundItem = {
  id: string;
  codeground_type: string;
  name: string;
  status: CodegroundStatus;
  updatedAt: string;
  createdAt: string;
  userId: number;
};

// A stopped codeground keeps its volume, so starting it again returns the same
// files at the same hostname — only its pod is gone.
const STATUS_LABEL: Record<CodegroundStatus, string> = {
  PROVISIONING: "Starting",
  RUNNING: "Running",
  STOPPED: "Stopped",
  FAILED: "Failed",
};

const STATUS_DOT: Record<CodegroundStatus, string> = {
  PROVISIONING: "bg-yellow-500",
  RUNNING: "bg-green-500",
  STOPPED: "bg-muted-foreground",
  FAILED: "bg-destructive",
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

  const setStatusInCache = (id: string, status: CodegroundStatus) => {
    queryClient.setQueryData(
      ["codegrounds"],
      (oldData: CodegroundItem[] | undefined) =>
        (oldData || []).map((cg) => (cg.id === id ? { ...cg, status } : cg)),
    );
  };

  const handleStart = async (id: string) => {
    setStatusInCache(id, "PROVISIONING");
    try {
      await api.post(`/codeground/${id}/start`);
      toast.success("Starting codeground");
    } catch {
      setStatusInCache(id, "STOPPED");
      toast.error("Failed to start codeground");
    }
  };

  const handleStop = async (id: string) => {
    try {
      await api.post(`/codeground/${id}/stop`);
      setStatusInCache(id, "STOPPED");
      toast.success("Stopping codeground; your files are kept");
    } catch {
      toast.error("Failed to stop codeground");
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
                onStart={handleStart}
                onStop={handleStop}
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
  onStart,
  onStop,
}: {
  codeground: CodegroundItem;
  onDelete: (id: string) => void;
  onStart: (id: string) => void;
  onStop: (id: string) => void;
}) => {
  const status = codeground.status ?? "PROVISIONING";
  const isRunning = status === "RUNNING";
  const isBusy = status === "PROVISIONING";

  // Opening a codeground that is not running would just fail to connect, so
  // the card is only a link once there is something to connect to.
  const header = (
    <>
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <FolderGit2 className="w-8 h-8 mb-2 text-primary" />
          <h2 className="font-semibold">{codeground.name}</h2>
        </div>
        <p className="text-xs text-muted-foreground">
          {codeground.codeground_type}
        </p>
      </div>
      <div className="flex items-center gap-2 mt-4">
        <span
          className={`inline-block w-2 h-2 rounded-full ${STATUS_DOT[status]}`}
          aria-hidden="true"
        />
        <span className="text-xs text-muted-foreground">
          {STATUS_LABEL[status]}
        </span>
        {codeground.updatedAt && (
          <span className="ml-auto text-xs text-muted-foreground">
            Updated {new Date(codeground.updatedAt).toLocaleDateString()}
          </span>
        )}
      </div>
    </>
  );

  return (
    <Card className="relative p-6 pb-14 transition-colors hover:border-primary">
      {isRunning ? (
        <Link href={`/codeground/${codeground.id}`} className="block">
          {header}
        </Link>
      ) : (
        <div className="opacity-70">{header}</div>
      )}

      <div className="absolute flex items-center gap-1 bottom-3 right-3">
        {isRunning ? (
          <Button
            variant="ghost"
            size="sm"
            className="h-7 gap-1 text-muted-foreground"
            onClick={() => onStop(codeground.id)}
          >
            <Square className="w-3 h-3" />
            Stop
          </Button>
        ) : (
          <Button
            variant="ghost"
            size="sm"
            className="h-7 gap-1 text-muted-foreground"
            disabled={isBusy}
            onClick={() => onStart(codeground.id)}
          >
            {isBusy ? (
              <Loader2 className="w-3 h-3 animate-spin" />
            ) : (
              <Play className="w-3 h-3" />
            )}
            Start
          </Button>
        )}
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
                This permanently deletes &quot;{codeground.name}&quot;, including
                its storage volume and every file in it. Stopping the codeground
                instead keeps your work. This action cannot be undone.
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
};

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
