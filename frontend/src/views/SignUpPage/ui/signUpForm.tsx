'use client';

import { signUp } from '../model/actions';
import { startTransition, useActionState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Link from 'next/link';

import { Card, CardContent } from '@/shared/ui/card';
import { FieldDescription, FieldGroup, FieldLabel, Field, FieldSeparator } from '@/shared/ui/field';
import { Input } from '@/shared/ui/input';
import { Button } from '@/shared/ui/button';

const signUpFormSchema = z.object({
  email: z.string().trim().email('Введите корректный email'),
  display_name: z.string().min(2, 'Имя должно содержать минимум 2 символа'),
  password: z.string().min(5, 'Пароль должен содержать минимум 5 символов'),
  avatar_url: z.string().optional(),
});

type SignUpFormValues = z.infer<typeof signUpFormSchema>;

export default function SignUpForm() {
  const router = useRouter();
  const [state, formAction, isPending] = useActionState(signUp, null);

  const {
    register,
    handleSubmit: handleFormSubmit,
    setError,
    formState: { errors, isValid },
  } = useForm<SignUpFormValues>({
    resolver: zodResolver(signUpFormSchema),
    mode: 'onChange',
    defaultValues: { email: '', display_name: '', password: '', avatar_url: '' },
  });

  useEffect(() => {
    if (state?.success) {
      router.push('/');
    }

    if (state?.errors) {
      Object.entries(state.errors).forEach(([key, messages]) => {
        setError(key as keyof SignUpFormValues, {
          type: 'server',
          message: Array.isArray(messages) ? messages[0] : (messages as string),
        });
      });
    }
  }, [state, router, setError]);

  const onSubmit = (data: SignUpFormValues) => {
    const formData = new FormData();
    formData.append('email', data.email);
    formData.append('display_name', data.display_name);
    formData.append('password', data.password);
    formData.append('avatar_url', data.avatar_url || '');

    startTransition(() => {
      formAction(formData);
    });
  };

  return (
    <div className="flex h-[calc(100vh-80px)] flex-col items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm md:max-w-4xl">
        <div className="flex flex-col gap-6">
          <Card className="overflow-hidden p-0">
            <CardContent className="grid p-0 md:grid-cols-2">
              <form onSubmit={handleFormSubmit(onSubmit)} className="p-6 md:p-8">
                <FieldGroup>
                  <div className="flex flex-col items-center gap-2 text-center">
                    <h1 className="text-2xl font-bold">Create an account</h1>
                    <p className="text-balance text-muted-foreground">
                      Sign up to start discovering and saving ideas
                    </p>
                  </div>

                  <Field>
                    <FieldLabel htmlFor="email">Email</FieldLabel>
                    <Input
                      id="email"
                      type="email"
                      placeholder="m@example.com"
                      {...register('email')}
                    />
                    {errors.email && (
                      <p className="text-xs text-destructive mt-1">{errors.email.message}</p>
                    )}
                  </Field>

                  <Field>
                    <FieldLabel htmlFor="display_name">Display Name</FieldLabel>
                    <Input
                      id="display_name"
                      type="text"
                      placeholder="Your name"
                      {...register('display_name')}
                    />
                    {errors.display_name && (
                      <p className="text-xs text-destructive mt-1">{errors.display_name.message}</p>
                    )}
                  </Field>

                  <Field>
                    <FieldLabel htmlFor="password">Password</FieldLabel>
                    <Input id="password" type="password" {...register('password')} />
                    {errors.password && (
                      <p className="text-xs text-destructive mt-1">{errors.password.message}</p>
                    )}
                  </Field>

                  <Field>
                    <FieldLabel htmlFor="avatar_url">Avatar URL (optional)</FieldLabel>
                    <Input
                      id="avatar_url"
                      type="text"
                      placeholder="https://example.com/avatar.jpg"
                      {...register('avatar_url')}
                    />
                    {errors.avatar_url && (
                      <p className="text-xs text-destructive mt-1">{errors.avatar_url.message}</p>
                    )}
                  </Field>

                  <Field>
                    <Button type="submit" disabled={isPending || !isValid}>
                      {isPending ? 'Creating account...' : 'Sign Up'}
                    </Button>
                  </Field>

                  {state?.message && !state?.success && (
                    <p className="text-sm text-destructive font-medium text-center mt-2">
                      {state.message}
                    </p>
                  )}

                  <FieldSeparator className="*:data-[slot=field-separator-content]:bg-card">
                    Or continue with
                  </FieldSeparator>

                  <Field className="grid grid-cols-3 gap-4">
                    <Button variant="outline" type="button">
                      <span className="sr-only">Sign up with Apple</span>
                    </Button>
                    <Button variant="outline" type="button">
                      <span className="sr-only">Sign up with Google</span>
                    </Button>
                    <Button variant="outline" type="button">
                      <span className="sr-only">Sign up with Meta</span>
                    </Button>
                  </Field>

                  <FieldDescription className="text-center">
                    Already have an account?{' '}
                    <Link
                      href="/signin"
                      className="underline underline-offset-4 hover:text-primary"
                    >
                      Sign in
                    </Link>
                  </FieldDescription>
                </FieldGroup>
              </form>

              <div className="relative hidden bg-muted md:block" />
            </CardContent>
          </Card>
          <FieldDescription className="px-6 text-center">
            By clicking continue, you agree to our{' '}
            <Link href="#" className="underline underline-offset-4 hover:text-primary">
              Terms of Service
            </Link>{' '}
            and{' '}
            <Link href="#" className="underline underline-offset-4 hover:text-primary">
              Privacy Policy
            </Link>
            .
          </FieldDescription>
        </div>
      </div>
    </div>
  );
}
