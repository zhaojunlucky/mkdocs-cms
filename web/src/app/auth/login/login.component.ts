import { Component, OnInit, ChangeDetectionStrategy } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { AuthService } from '../auth.service';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatDividerModule } from '@angular/material/divider';

import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import {PageTitleService} from '../../services/page.title.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatDividerModule,
    MatProgressSpinnerModule,
    MatSnackBarModule
],
  templateUrl: './login.component.html',
  changeDetection: ChangeDetectionStrategy.Eager,
  styleUrls: ['./login.component.scss']
})
export class LoginComponent implements OnInit {
  loading = false;

  constructor(
    private readonly authService: AuthService,
    private readonly route: ActivatedRoute,
    private readonly snackBar: MatSnackBar,
    private readonly pageTitleService: PageTitleService
  ) {}

  ngOnInit(): void {
    this.pageTitleService.title = 'Login';
    const initialReturnUrl = this.route.snapshot.queryParamMap.get('returnUrl');
    this.authService.rememberReturnUrl(initialReturnUrl);
    // Check if user is already logged in
    if (this.authService.isLoggedIn) {
      this.route.queryParams.subscribe(params => {
        this.authService.navigateAfterLogin(params['returnUrl']);
      })
      return;
    }

    // Handle OAuth callback
    this.route.queryParams.subscribe(params => {
      const token = params['token'];

      if (token) {
        this.loading = true;
        const searchParams = new URLSearchParams();
        searchParams.set('token', token);

        this.authService.handleAuthCallback(searchParams).subscribe({
          next: () => {
            this.loading = false;
            this.authService.navigateAfterLogin(this.route.snapshot.queryParamMap.get('returnUrl'));
          },
          error: (error) => {
            this.loading = false;
            console.error('Authentication error:', error);
            this.snackBar.open('Authentication failed. Please try again.', 'Close', {
              duration: 5000,
              panelClass: 'error-snackbar'
            });
          }
        });
      }
    });
  }

  loginWithGithub(): void {
    this.loading = true;
    this.authService.loginWithGithub();
  }

  loginWithGoogle(): void {
    this.loading = true;
    this.authService.loginWithGoogle();
  }
}
