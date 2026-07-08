// app/actions.ts
"use server";

import { z } from "zod";

// 1. Define the same schema on the server for double-validation (optional but recommended)
const formSchema = z.object({
  username: z.string().min(3),
  email: z.string().email(),
});

export async function processForm(data: z.infer<typeof formSchema>) {
  // 2. Validate incoming data
  const parsed = formSchema.safeParse(data);
  
  if (!parsed.success) {
    return { error: "Invalid data submitted" };
  }

  // 3. Perform your database operations here
  console.log("Saving to database:", parsed.data);

  // 4. Return success state or revalidate paths
  return { success: true, message: `Welcome, ${parsed.data.username}!` };
}