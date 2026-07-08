'use client';

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import Link from 'next/link';
import processFormAction, { SignInFormData } from '@/shared/server-actions/processFormAction'; // FSD location
import { signInSchema } from '@/entities/user/schema';

export function LoginForm() {
  const formMethods = useForm<SignInFormData>({ resolver: zodResolver(signInSchema) });

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 max-w-md p-6 bg-card rounded-lg border shadow-sm"> {/* Username Field */}
      </div> // Email field