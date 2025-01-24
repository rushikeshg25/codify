"use client";

import NewPlayground from "@/components/new-playground";
import { Card } from "@/components/ui/card";
import { FolderGit2 } from "lucide-react";
import Link from "next/link";

export default function DashboardPage() {
  const codegrounds = [
    {
      id: 1,
      name: "React Todo App",
      language: "TypeScript",
      lastEdited: "2 hours ago",
    },
    { id: 2, name: "Python API", language: "Python", lastEdited: "1 day ago" },
    {
      id: 3,
      name: "Portfolio Website",
      language: "JavaScript",
      lastEdited: "3 days ago",
    },
  ];

  return (
    <div className="min-h-screen bg-background p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold">My Codegrounds</h1>
          <NewPlayground />
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {codegrounds.map((codeground) => (
            <Link href={`/codeground/${codeground.id}`} key={codeground.id}>
              <Card className="p-6 hover:border-primary transition-colors cursor-pointer">
                <div className="flex items-start justify-between">
                  <div className="space-y-1">
                    <FolderGit2 className="w-8 h-8 text-primary mb-2" />
                    <h2 className="font-semibold">{codeground.name}</h2>
                    <p className="text-sm text-muted-foreground">
                      {codeground.language}
                    </p>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {codeground.lastEdited}
                  </p>
                </div>
              </Card>
            </Link>
          ))}
        </div>
      </div>
    </div>
  );
}
