"use server";

import { z } from 'zod';
import { cookies } from "next/headers";

export const signUpSchema = z.object({
  username: z.string().min(3, "Username must be at least 3 characters long."),
  email: z.string().trim().email("Invalid email format."),
  password: z.string().min(8, "Password must be at least 8 characters long."),
});

export async function signUpAction(formData: { username: string; email: string; password: string }) {
  const result = signUpSchema.safeParse({
    username: formData.username,
    email: formData.email,
    password: formData.password
  });

  if (!result.success) {
    return { success: false, error: "Validation failed.", details: result.error.issues };
  }

  const { username, email, password } = result.data;

  console.log(`Attempting registration for ${username} (${email})...`);

  /*
  try {
    const response = await authService.register(username, email, password);
    if (response?.user) {
      // Optionally log the user in immediately upon successful registration
      cookies().set("session_token", response.token, { httpOnly: true, secure: process.env.NODE_ENV === 'production', maxAge: 60 * 60 * 24 });
      return { success: true };
    } else {
      // Handle API specific failure (e.g., email already exists)
      return { success: false, error: "Registration failed. User might exist or general server issue." };
    }
  } catch (error) {
    console.error("Signup API Error:", error);
    return { success: false, error: "A server error occurred during registration." };
  }
  */

  // MOCK SUCCESS/FAILURE FOR DEMONSTRATION PURPOSES
  if (email.includes('@')) { // Simple check to simulate successful validation and creation
      console.log(`Successfully registered mock user ${username}`);
      return { success: true };
  } else {
      return { success: false, error: "Registration failed due to invalid data." };
  }
}
