/** Успешный ответ API: обёртка `{"data": ...}`. */
export type ApiResponse<T> = {
  data: T;
};

/** Ошибочный ответ API: обёртка `{"error": {...}}`. */
export type ApiError = {
  error: {
    code: string;
    message: string;
    details?: string;
  };
};

export type ApiPaginatedResponse<T> = {
  data: T[];
  next_cursor: string | null;
};