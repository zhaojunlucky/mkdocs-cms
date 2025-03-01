import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { RepositoryService, Repository } from '../services/repository.service';
import { AuthService } from '../auth/auth.service';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.scss']
})
export class HomeComponent implements OnInit {
  repositories: Repository[] = [];
  loading = true;
  error = '';
  
  constructor(
    private repositoryService: RepositoryService,
    private authService: AuthService
  ) {}
  
  ngOnInit(): void {
    this.loadRepositories();
  }
  
  loadRepositories(): void {
    this.loading = true;
    this.error = '';
    
    this.repositoryService.getUserRepositories().subscribe({
      next: (repos) => {
        this.repositories = repos;
        this.loading = false;
      },
      error: (err) => {
        console.error('Error loading repositories:', err);
        this.error = 'Failed to load repositories. Please try again later.';
        this.loading = false;
      }
    });
  }
}
