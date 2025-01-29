"use client";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
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

const NewCodeground = () => {
  const [name, setName] = useState("");
  const [type, setType] = useState("");

  const submitHandler = async () => {
    if (!name || !type) {
      return toast.error("Please fill all the fields");
    }
    const res = await api.post("/codeground", {
      name,
      codeground_type: type,
    });
    console.log(res);
    toast.success("Codeground created successfully");
  };
  return (
    <AlertDialog>
      <AlertDialogTrigger asChild>
        <Button className="gap-2">
          <Plus className="w-4 h-4" />
          New Codeground
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>New Codeground</AlertDialogTitle>
          <AlertDialogDescription className="flex flex-col gap-5">
            <div>
              <div className="grid w-full items-center gap-4">
                <div className="flex flex-col space-y-1.5">
                  <Label htmlFor="name">Name</Label>
                  <Input
                    onChange={(e) => setName(e.target.value)}
                    placeholder="Name your Codeground"
                  />
                </div>
                <div className="flex flex-col space-y-1.5">
                  <Label htmlFor="type">Type</Label>
                  <Select
                    onValueChange={(value) => setType(value)}
                    value={type}
                  >
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
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={submitHandler}>Create</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
};

export default NewCodeground;
