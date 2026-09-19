"use client";

import { useSearchParams } from "next/navigation";
import Workspace, { WorkspaceCodeground } from "@/components/codeground/Workspace";

export default function CodegroundClient() {
  const searchParams = useSearchParams();
  const data = searchParams.get("data");
  const codeground: WorkspaceCodeground | null = data ? JSON.parse(data) : null;

  if (!codeground) return <div>No Codeground Data Found</div>;

  return <Workspace codeground={codeground} />;
}
