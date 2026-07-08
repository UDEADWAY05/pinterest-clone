'use client';

import { login } from '../model/actions';
import { startTransition, useActionState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

import { Card, CardContent } from '@/shared/ui/card';
import { FieldDescription, FieldGroup, FieldLabel, Field, FieldSeparator } from '@/shared/ui/field';
import { Input } from '@/shared/ui/input';
import { Button } from '@/shared/ui/button';

// 1. Описываем схему валидации для фронтенда
const loginSchema = z.object({
  email: z.string().email('Введите корректный email адрес'),
  password: z.string().min(5, 'Пароль должен содержать минимум 5 символов'),
});

// Выводим тип данных из схемы
type LoginFormValues = z.infer<typeof loginSchema>;

export function LoginPage() {
  const router = useRouter();
  const [state, formAction, isPending] = useActionState(login, null);

  const {
    register,
    handleSubmit: handleFormSubmit,
    setError,
    formState: { errors, isValid },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    mode: 'onChange',
    defaultValues: { email: '', password: '' },
  });

  useEffect(() => {
    if (state?.success) {
      router.push('/');
    }

    // Если бэкенд (Go) или Zod на сервере вернул ошибки конкретных полей
    if (state?.errors) {
      Object.entries(state.errors).forEach(([key, messages]) => {
        setError(key as keyof LoginFormValues, {
          type: 'server',
          message: Array.isArray(messages) ? messages[0] : messages,
        });
      });
    }
  }, [state, router, setError]);

  // 2. Обновляем функцию отправки
  const onSubmit = (data: LoginFormValues) => {
    const formData = new FormData();
    formData.append('email', data.email);
    formData.append('password', data.password);

    // Оборачиваем вызов в startTransition, чтобы React корректно трекал статус isPending
    startTransition(() => {
      formAction(formData);
    });
  };
  // 4. Функция отправки формы
  return (
    <div className="flex h-[calc(100vh-80px)] flex-col items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm md:max-w-4xl">
        <div className="flex flex-col gap-6">
          <Card className="overflow-hidden p-0">
            <CardContent className="grid p-0 md:grid-cols-2">
              {/* Передаем handleFormSubmit из RHF */}
              <form onSubmit={handleFormSubmit(onSubmit)} className="p-6 md:p-8">
                <FieldGroup>
                  <div className="flex flex-col items-center gap-2 text-center">
                    <h1 className="text-2xl font-bold">Welcome back</h1>
                    <p className="text-balance text-muted-foreground">
                      Login to your Acme Inc account
                    </p>
                  </div>

                  <Field>
                    <FieldLabel htmlFor="email">Email</FieldLabel>
                    {/* Регистрируем поле в RHF и не забываем про атрибут name */}
                    <Input
                      id="email"
                      type="email"
                      placeholder="m@example.com"
                      {...register('email')}
                    />
                    {/* Вывод ошибки email */}
                    {errors.email && (
                      <p className="text-xs text-destructive mt-1">{errors.email.message}</p>
                    )}
                  </Field>

                  <Field>
                    <div className="flex items-center">
                      <FieldLabel htmlFor="password">Password</FieldLabel>
                      <a href="#" className="ml-auto text-sm underline-offset-2 hover:underline">
                        Forgot your password?
                      </a>
                    </div>
                    {/* Регистрируем поле в RHF */}
                    <Input id="password" type="password" {...register('password')} />
                    {/* Вывод ошибки пароля */}
                    {errors.password && (
                      <p className="text-xs text-destructive mt-1">{errors.password.message}</p>
                    )}
                  </Field>

                  <Field>
                    {/* Блокируем кнопку при отправке или если форма невалидна на клиенте */}
                    <Button type="submit" disabled={isPending || !isValid}>
                      {isPending ? 'Signing in...' : 'Login'}
                    </Button>
                  </Field>

                  {/* Вывод глобальной ошибки (например, 401 от Go бэкенда) */}
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
                      {/* SVG Apple */}
                      <span className="sr-only">Login with Apple</span>
                    </Button>
                    <Button variant="outline" type="button">
                      {/* SVG Google */}
                      <span className="sr-only">Login with Google</span>
                    </Button>
                    <Button variant="outline" type="button">
                      {/* SVG Meta */}
                      <span className="sr-only">Login with Meta</span>
                    </Button>
                  </Field>

                  <FieldDescription className="text-center">
                    Don&apos;t have an account? <a href="#">Sign up</a>
                  </FieldDescription>
                </FieldGroup>
              </form>

              <div className="relative hidden bg-muted md:block"></div>
            </CardContent>
          </Card>
          <FieldDescription className="px-6 text-center">
            By clicking continue, you agree to our <a href="#">Terms of Service</a> and{' '}
            <a href="#">Privacy Policy</a>.
          </FieldDescription>
        </div>
      </div>
    </div>
  );
}
