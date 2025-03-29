export interface ArrayResponse<T> {
  entryCount: number;
  entries: T[];
}

export interface ErrorMessage {
  message: string;
}

export interface ErrorResponse {
  httpStatusCode: number;
  errorMessages: ErrorMessage[]
}
