"use client";
import { Button } from "@/components/ui/button";
import { Loader2, Plus } from "lucide-react";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useState } from "react";
import toast from "react-hot-toast";
import api from "@/lib/api";
import { codeground } from "@/types/codeground";

interface NewCodegroundProps {
  onCodegroundCreated: (codeground: codeground) => void;
}

const NewCodeground = ({ onCodegroundCreated }: NewCodegroundProps) => {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [type, setType] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const submitHandler = async () => {
    if (!name || !type) {
      toast.error("Please fill all the fields");
      return;
    }
    setSubmitting(true);
    try {
      const res = await api.post("/codeground", {
        name,
        codeground_type: type,
      });
      onCodegroundCreated(res.data);
      toast.success("Codeground created successfully");
      setName("");
      setType("");
      setOpen(false);
    } catch {
      toast.error("Failed to create codeground");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger asChild>
        <Button className="gap-1">
          <Plus className="w-3 h-3" />
          New Codeground
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>New Codeground</AlertDialogTitle>
          <AlertDialogDescription>
            Name your codeground and pick a template to get started.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <div className="grid items-center w-full gap-4">
          <div className="flex flex-col space-y-1.5">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Name your Codeground"
            />
          </div>
          <div className="flex flex-col space-y-1.5">
            <Label htmlFor="type">Type</Label>
            <Select onValueChange={(value) => setType(value)} value={type}>
              <SelectTrigger id="type">
                <SelectValue placeholder="Select" />
              </SelectTrigger>
              <SelectContent position="popper">
                <SelectItem value="NODE">Node</SelectItem>
                <SelectItem value="REACT">React</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={submitting}>Cancel</AlertDialogCancel>
          <Button onClick={submitHandler} disabled={submitting} className="gap-1">
            {submitting && <Loader2 className="w-4 h-4 animate-spin" />}
            Create
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
};

export default NewCodeground;
