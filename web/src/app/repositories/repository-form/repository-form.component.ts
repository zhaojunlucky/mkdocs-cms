import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { RepositoryService } from '../../services/repository.service';
import { AuthService } from '../../auth/auth.service';

@Component({
  selector: 'app-repository-form',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './repository-form.component.html',
  styleUrls: ['./repository-form.component.scss']
})
export class RepositoryFormComponent implements OnInit {
  repoForm!: FormGroup;
  loading = false;
  error = '';
  
  constructor(
    private fb: FormBuilder,
    private router: Router,
    private repositoryService: RepositoryService,
    private authService: AuthService
  ) {}
  
  ngOnInit(): void {
    this.initForm();
  }
  
  initForm(): void {
    this.repoForm = this.fb.group({
      name: ['', [Validators.required]],
      url: ['', [Validators.required]],
      local_path: ['', [Validators.required]]
    });
  }
  
  onSubmit(): void {
    if (this.repoForm.invalid) {
      return;
    }
    
    this.loading = true;
    this.error = '';
    
    const userId = this.authService.currentUser?.id;
    if (!userId) {
      this.error = 'User not authenticated';
      this.loading = false;
      return;
    }
    
    const repoData = {
      ...this.repoForm.value,
      user_id: userId
    };
    
    this.repositoryService.createRepository(repoData).subscribe({
      next: (repo) => {
        this.loading = false;
        this.router.navigate(['/repositories', repo.id]);
      },
      error: (err) => {
        console.error('Error creating repository:', err);
        this.error = 'Failed to create repository. Please try again later.';
        this.loading = false;
      }
    });
  }
}
