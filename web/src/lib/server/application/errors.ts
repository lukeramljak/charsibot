export type ApplicationErrorCode =
  | 'conflict'
  | 'forbidden'
  | 'invalid_input'
  | 'not_found'
  | 'not_ready'
  | 'unavailable';

export class ApplicationError extends Error {
  readonly code: ApplicationErrorCode;

  constructor(code: ApplicationErrorCode, message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = 'ApplicationError';
    this.code = code;
  }
}
