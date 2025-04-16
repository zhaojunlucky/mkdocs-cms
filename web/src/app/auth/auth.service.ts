import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { BehaviorSubject, Observable, of } from 'rxjs';
import { HttpClient } from '@angular/common/http';
import { catchError, map, tap } from 'rxjs/operators';
import {environment} from '../../environments/environment';

export interface User {
  id: string;
  username: string;
  name: string;
  email: string;
  avatar_url?: string;
  provider: 'github' | 'google';
}

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private userSubject = new BehaviorSubject<User | null>(null);
  private tokenSubject = new BehaviorSubject<string | null>(null);
  private apiUrl = environment.apiServer; // Base URL for our backend API

  constructor(
    private http: HttpClient,
    private router: Router
  ) {
    this.refreshToken();
  }

  get user(): Observable<User | null> {
    return this.userSubject.asObservable();
  }

  get token(): Observable<string | null> {
    return this.tokenSubject.asObservable();
  }

  get currentToken(): string | null {
    this.refreshToken();
    return this.tokenSubject.value;
  }

  get currentUser(): User | null {
    return this.userSubject.value;
  }

  get isLoggedIn(): boolean {
    const token = this.tokenSubject.value;
    if (!this.userSubject.value || !token) {
      return false;
    }

    try {
      const tokenParts = token.split('.');
      if (tokenParts.length !== 3) {
        return false;
      }

      const payload = JSON.parse(atob(tokenParts[1]));
      const expirationTime = payload.exp * 1000; // Convert to milliseconds
      return Date.now() < expirationTime;
    } catch {
      return false;
    }
  }

  loginWithGithub(): void {
    // Redirect to GitHub OAuth login page
    window.location.href = `${this.apiUrl}/auth/github`;
  }

  loginWithGoogle(): void {
    // Redirect to Google OAuth login page
    window.location.href = `${this.apiUrl}/auth/google`;
  }

  // Handle OAuth callback
  handleAuthCallback(params: URLSearchParams): Observable<User> {
    const token = params.get('token');
    if (!token) {
      throw new Error('Authentication failed');
    }

    // Store token
    localStorage.setItem('auth_token', token);
    this.tokenSubject.next(token);

    // Get user info
    return this.getUserInfo(token).pipe(
      tap(user => this.setUser(user)),
      catchError(error => {
        this.logout();
        throw error;
      })
    );
  }

  // Get user info from token
  getUserInfo(token: string): Observable<User> {
    return this.http.get<User>(`${this.apiUrl}/auth/user`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
  }

  // Check if token is valid and refresh user info
  checkAuthStatus(): Observable<boolean> {
    const token = this.currentToken;
    if (!token) {
      return of(false);
    }

    return this.getUserInfo(token).pipe(
      map(user => {
        this.setUser(user);
        return true;
      }),
      catchError(() => {
        this.logout();
        return of(false);
      })
    );
  }

  // Store user in localStorage and update subject
  setUser(user: User): void {
    localStorage.setItem('user', JSON.stringify(user));
    this.userSubject.next(user);
    this.router.navigate(['/home']);
  }

  logout(): void {
    localStorage.removeItem('user');
    localStorage.removeItem('auth_token');
    this.userSubject.next(null);
    this.tokenSubject.next(null);
    this.router.navigate(['/login']);
  }

  private refreshToken() {
    // Check if user is already logged in (from localStorage)
    const storedUser = localStorage.getItem('user');
    const storedToken = localStorage.getItem('auth_token');

    if (storedUser && storedToken) {
      this.userSubject.next(JSON.parse(storedUser));
      this.tokenSubject.next(storedToken);
    } else {
      this.userSubject.next(null);
      this.tokenSubject.next(null);
    }
  }
}
