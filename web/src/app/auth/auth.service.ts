import { Injectable } from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {BehaviorSubject, Observable, of, shareReplay} from 'rxjs';
import { HttpClient } from '@angular/common/http';
import { catchError, map, tap } from 'rxjs/operators';
import {environment} from '../../environments/environment';
import {StrUtils} from '../shared/utils/str.utils';

export interface User {
  id: string;
  username: string;
  name: string;
  email: string;
  avatar_url?: string;
  provider: 'github' | 'google';
  expires_at: number;
}

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private userSubject = new BehaviorSubject<User | null>(null);
  private apiUrl = environment.apiServer; // Base URL for our backend API

  constructor(
    private http: HttpClient,
    private router: Router,
    private route: ActivatedRoute
  ) {
    this.refreshUserAuth()
  }

  refreshUserAuth() {
      this.getUserInfo().pipe(
        shareReplay()
      ).subscribe( {
          next: (user: User) => {
            this.setUser(user, null);
            if (this.router.url.startsWith('/login')) {
              let returnUrl = this.route.snapshot.queryParamMap.get('returnUrl');
              if (returnUrl) {
                this.router.navigate([returnUrl]);
              } else {
                this.router.navigate(['/home']);
              }
            }
          },
          error: (err) => {
            this.logout();
            throw err
          }
        }

      )

  }

  get user(): Observable<User | null> {
    return this.userSubject.asObservable();
  }


  get currentUser(): User | null {
    return this.userSubject.value;
  }

  get isLoggedIn(): boolean {
    if (!this.userSubject.value) {
      return false;
    }
    const expirationTime = this.userSubject.value.expires_at * 1000; // Convert to milliseconds
    return Date.now() < expirationTime;
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
    // token is user id sha256

    // Get user info
    return this.getUserInfo().pipe(
      tap(user => {
        this.setUser(user, token);
        this.router.navigate(['/home']);
      }),
      catchError(error => {
        this.logout();
        throw error;
      })
    );
  }

  // Get user info from token
  getUserInfo(): Observable<User> {
    return this.http.get<User>(`${this.apiUrl}/auth/user`);
  }

  // Check if token is valid and refresh user info
  checkAuthStatus(): Observable<boolean> {
    if (!this.userSubject.value) {
      return of(false);
    }

    return this.getUserInfo().pipe(
      map(user => {
        this.setUser(user, null);
        return true;
      }),
      catchError(() => {
        this.logout();
        return of(false);
      })
    );
  }

  // Store user in localStorage and update subject
  setUser(user: User, token: string|null): void {
    // validate token
    sessionStorage.setItem('user', JSON.stringify(user));
    this.userSubject.next(user);

  }

  logout(): void {
    this.http.delete(`${this.apiUrl}/auth/logout`).subscribe({
      next: () => {
        sessionStorage.removeItem('user');
        this.userSubject.next(null);
        this.router.navigate(['/login']);
      },
      error: (error) => {
        console.error('Error logging out:', error);
        alert(StrUtils.stringifyHTTPErr(error));
      }
    });

  }
}
