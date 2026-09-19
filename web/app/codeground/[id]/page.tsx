import { Suspense } from "react";
import CodegroundClient from "@/components/codeground/CodegroundClient";
import WorkspaceSkeleton from "@/components/codeground/WorkspaceSkeleton";

export default function CodegroundPage() {
  return (
    <Suspense fallback={<WorkspaceSkeleton />}>
      <CodegroundClient />
    </Suspense>
  );
}
