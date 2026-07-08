import * as z from 'zod';

/** Схемы валидации для авторизации */
export const signInSchema = z.object({
  username: z.string().min(3, "Имя пользователя должно быть минимум из 3 символов."),
  password: z.string().min(8, "Пароль должен быть минимум 8 символов.").optional(), // optional для процессоров форм
});

export const signUpSchema = z.object({
  username: z.string().min(3, "Имя пользователя должно быть минимум из 3 символов."),
  email: z.string().trim().email("Некорректный формат электронной почты"),
  password: z.string().min(8, "Пароль должен быть минимум 8 символов.").optional(), // optional для процессоров форм
});

/** Типы данных из схем */
export type SignInFormData = z.infer<typeof signInSchema>;
export type SignUpFormData = z.infer<typeof signUpSchema>;
export type AuthFormData = SignInFormData | SignUpFormData;
