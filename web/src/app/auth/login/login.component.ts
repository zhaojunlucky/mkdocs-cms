import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { AuthService } from '../auth.service';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatDividerModule } from '@angular/material/divider';
import { CommonModule } from '@angular/common';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatDividerModule,
    MatProgressSpinnerModule,
    MatSnackBarModule
  ],
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent implements OnInit {
  loading = false;
  
  constructor(
    private authService: AuthService,
    private route: ActivatedRoute,
    private router: Router,
    private snackBar: MatSnackBar
  ) {}

  ngOnInit(): void {
    // Check if user is already logged in
    if (this.authService.isLoggedIn) {
      this.router.navigate(['/home']);
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
            this.router.navigate(['/home']);
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
