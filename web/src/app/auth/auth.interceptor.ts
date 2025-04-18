import { HttpInterceptorFn } from '@angular/common/http';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  // const authService = inject(AuthService);

  // Only add token for API requests
  if (req.url.includes('/api/')) {
    const authReq = req.clone({
      withCredentials: true,
    });
    return next(authReq)
  }

  return next(req);
};
