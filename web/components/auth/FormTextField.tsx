"use client";

import { Control, FieldPath, FieldValues } from "react-hook-form";
import { Input } from "@/components/ui/input";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";

interface FormTextFieldProps<T extends FieldValues> {
  control: Control<T>;
  name: FieldPath<T>;
  label: string;
  placeholder: string;
  type?: React.ComponentProps<typeof Input>["type"];
}

export function FormTextField<T extends FieldValues>({
  control,
  name,
  label,
  placeholder,
  type,
}: FormTextFieldProps<T>) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => (
        <FormItem>
          <FormLabel>{label}</FormLabel>
          <FormControl>
            <Input placeholder={placeholder} type={type} {...field} />
          </FormControl>
          <FormMessage />
        </FormItem>
      )}
    />
  );
}
