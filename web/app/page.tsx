"use client";

import { Button } from "@/components/ui/button";
import { Terminal, Code2, Github } from "lucide-react";
import Link from "next/link";

export default function Home() {
  return (
    <div className="min-h-screen bg-background">
      {/* Hero Section */}
      <div className="container mx-auto px-4 pt-20 pb-16">
        <div className="text-center space-y-6">
          <h1 className="text-6xl font-bold tracking-tighter">
            Code Anywhere, <span className="text-primary">Together</span>
          </h1>
          <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
            A powerful, collaborative coding environment in your browser. Build,
            test, and deploy your projects with ease.
          </p>
          <div className="flex gap-4 justify-center pt-4">
            <Link href="/signup">
              <Button size="lg" className="gap-2">
                <Code2 className="w-4 h-4" />
                Start Coding Now
              </Button>
            </Link>
          </div>
        </div>

        {/* Features */}
        <div className="grid md:grid-cols-3 gap-8 mt-20">
          <div className="p-6 rounded-lg border bg-card">
            <Terminal className="w-12 h-12 text-primary mb-4" />
            <h3 className="text-xl font-semibold mb-2">Integrated Terminal</h3>
            <p className="text-muted-foreground">
              Full-featured terminal access for running commands, managing
              packages, and more.
            </p>
          </div>
          <div className="p-6 rounded-lg border bg-card">
            <Code2 className="w-12 h-12 text-primary mb-4" />
            <h3 className="text-xl font-semibold mb-2">VS Code Experience</h3>
            <p className="text-muted-foreground">
              Familiar VS Code-like interface with syntax highlighting and
              powerful editing features.
            </p>
          </div>
          <div className="p-6 rounded-lg border bg-card">
            <Github className="w-12 h-12 text-primary mb-4" />
            <h3 className="text-xl font-semibold mb-2">Git Integration</h3>
            <p className="text-muted-foreground">
              Seamless Git integration for version control and collaboration.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
