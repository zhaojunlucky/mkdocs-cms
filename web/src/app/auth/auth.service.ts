import { Injectable } from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {BehaviorSubject, from, Observable, of, shareReplay} from 'rxjs';
import { HttpClient } from '@angular/common/http';
import { catchError, map, switchMap } from 'rxjs/operators';
import {environment} from '../../environments/environment';
import {StrUtils} from '../shared/utils/str.utils';
import {MatDialog} from '@angular/material/dialog';
import {ConfirmDialogComponent} from '../shared/dialogs/confirm-dialog.component';

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
  private readonly returnUrlStorageKey = 'mkdocs-cms.returnUrl';

  constructor(
    private http: HttpClient,
    private router: Router,
    private route: ActivatedRoute,
    private dialog: MatDialog
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
              this.navigateAfterLogin(this.route.snapshot.queryParamMap.get('returnUrl'));
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
      switchMap(user => from(this.setUser(user, token)).pipe(map(() => user))),
      catchError(error => {
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
  async setUser(user: User, token: string | null): Promise<void> {
    if (token ) {
      const userIdHash = await StrUtils.sha256(user.id);
      if (userIdHash !== token) {
        this.showMessage('Authentication failed', 'Authentication failed, invalid token. Please try again.');
        throw new Error('Authentication failed');
      }

    }
    sessionStorage.setItem('user', JSON.stringify(user));
    this.userSubject.next(user);

  }

  rememberReturnUrl(returnUrl: string | null | undefined): void {
    if (returnUrl && !returnUrl.startsWith('/login') && !returnUrl.startsWith('/error')) {
      sessionStorage.setItem(this.returnUrlStorageKey, returnUrl);
    }
  }

  navigateAfterLogin(returnUrl?: string | null): void {
    const target = returnUrl || sessionStorage.getItem(this.returnUrlStorageKey);
    sessionStorage.removeItem(this.returnUrlStorageKey);
    if (target) {
      let routeParams = StrUtils.parseRedirectUrl(target);
      this.router.navigate(routeParams['paths'], { queryParams: routeParams['queryParams'] });
      return;
    }
    this.router.navigate(['/home']);
  }

  logout(): void {
    this.http.delete(`${this.apiUrl}/auth/logout`).subscribe({
      next: () => {
        sessionStorage.removeItem('user');
        this.userSubject.next(null);
        if (!this.router.url.startsWith('/error')) {
          this.router.navigate(['/login']);
        }
      },
      error: (error) => {
        console.error('Error logging out:', error);
        this.showMessage('Logout failed', StrUtils.stringifyHTTPErr(error), true);
      }
    });

  }

  private showMessage(title: string, message: string, destructive = false): void {
    this.dialog.open(ConfirmDialogComponent, {
      data: {
        title,
        message,
        confirmLabel: 'OK',
        cancelLabel: null,
        destructive,
        icon: destructive ? 'error' : 'info'
      }
    });
  }
}
