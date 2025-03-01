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

  toggleMenu(repo: Repository): void {
    // Close all other menus first
    this.repositories.forEach(r => {
      if (r !== repo) {
        r.showMenu = false;
      }
    });
    
    // Toggle the menu for the selected repository
    repo.showMenu = !repo.showMenu;
  }

  syncRepository(repo: Repository): void {
    repo.showMenu = false; // Close the menu
    repo.syncing = true; // Show syncing indicator
    
    this.repositoryService.syncRepository(repo.id).subscribe({
      next: () => {
        // Reload the repositories to get updated status
        this.loadRepositories();
      },
      error: (err) => {
        console.error('Error syncing repository:', err);
        repo.syncing = false;
        // Optionally show an error message
      }
    });
  }

  editRepository(repo: Repository): void {
    repo.showMenu = false; // Close the menu
    // Implement edit functionality or navigation
    console.log('Edit repository:', repo.id);
  }

  deleteRepository(repo: Repository): void {
    repo.showMenu = false; // Close the menu
    if (confirm(`Are you sure you want to delete repository "${repo.name}"?`)) {
      this.repositoryService.deleteRepository(repo.id).subscribe({
        next: () => {
          this.repositories = this.repositories.filter(r => r.id !== repo.id);
        },
        error: (err) => {
          console.error('Error deleting repository:', err);
          // Optionally show an error message
        }
      });
    }
  }

  getStatusClass(repo: Repository): string {
    if (repo.syncing) {
      return 'status-syncing';
    }
    
    switch (repo.status) {
      case 'synced':
        return 'status-synced';
      case 'failed':
        return 'status-failed';
      case 'pending':
        return 'status-pending';
      default:
        return '';
    }
  }
}
