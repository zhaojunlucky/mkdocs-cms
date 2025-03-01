import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { RepositoryService, Repository, SyncResponse, AsyncTask } from '../services/repository.service';
import { AuthService } from '../auth/auth.service';
import { interval } from 'rxjs';
import { switchMap, takeWhile } from 'rxjs/operators';
import { ComponentsModule } from '../components/components.module';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, RouterModule, ComponentsModule, MatButtonModule, MatIconModule, MatMenuModule],
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.scss']
})
export class HomeComponent implements OnInit {
  repositories: Repository[] = [];
  loading = true;
  error = '';
  activeTasks: Map<string, AsyncTask> = new Map(); // Map of repo ID to task
  
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
      error: (err: any) => {
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
      next: (response: SyncResponse) => {
        console.log('Sync started with task ID:', response.task_id);
        
        // Create a temporary AsyncTask object until we get the full details
        const tempTask: AsyncTask = {
          id: response.task_id,
          type: 'sync',
          status: 'pending',
          resource_id: repo.id.toString(),
          user_id: '',
          message: 'Sync in progress...',
          progress: 0,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        };
        
        // Start polling for task status
        this.activeTasks.set(repo.id.toString(), tempTask);
        this.pollTaskStatus(response.task_id, repo);
      },
      error: (err: any) => {
        console.error('Error starting repository sync:', err);
        repo.syncing = false;
        // Optionally show an error message
      }
    });
  }

  // Poll for task status updates
  private pollTaskStatus(taskId: string, repo: Repository, intervalMs: number = 2000): void {
    const subscription = interval(intervalMs).pipe(
      switchMap(() => this.repositoryService.getTask(taskId)),
      takeWhile((task: AsyncTask) => task.status === 'pending' || task.status === 'running', true)
    ).subscribe({
      next: (task: AsyncTask) => {
        console.log('Task status:', task.status);
        
        if (task.status === 'completed' || task.status === 'failed') {
          // Task is done, stop polling and reload repositories
          subscription.unsubscribe();
          repo.syncing = false;
          this.activeTasks.delete(repo.id.toString());
          this.loadRepositories();
        } else {
          this.activeTasks.set(repo.id.toString(), task);
        }
      },
      error: (err: any) => {
        console.error('Error polling task status:', err);
        subscription.unsubscribe();
        repo.syncing = false;
        this.activeTasks.delete(repo.id.toString());
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
        error: (err: any) => {
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
