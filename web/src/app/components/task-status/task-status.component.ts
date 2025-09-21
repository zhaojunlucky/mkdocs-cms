import { Component, Input, OnInit, OnDestroy } from '@angular/core';
import { AsyncTask, RepositoryService, Task } from '../../services/repository.service';
import { interval, Subscription } from 'rxjs';
import { switchMap, takeWhile } from 'rxjs/operators';
import { NgClass, TitleCasePipe } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatProgressBarModule } from '@angular/material/progress-bar';

@Component({
  selector: 'app-task-status',
  templateUrl: './task-status.component.html',
  imports: [
    NgClass,
    TitleCasePipe,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatListModule,
    MatChipsModule,
    MatProgressSpinnerModule,
    MatProgressBarModule
  ],
  styleUrls: ['./task-status.component.scss']
})
export class TaskStatusComponent implements OnInit, OnDestroy {
  @Input() taskId: string | undefined;
  @Input() task: Task | AsyncTask | undefined;
  @Input() autoRefresh = true;
  @Input() refreshInterval = 2000; // 2 seconds

  asyncTask: AsyncTask | null = null;
  loading = true;
  error = '';
  private subscription: Subscription | null = null;

  constructor(private repositoryService: RepositoryService) {}

  ngOnInit(): void {
    if (this.taskId) {
      this.loadTask();
      if (this.autoRefresh) {
        this.startPolling();
      }
    } else if (this.task) {
      if ('type' in this.task && 'resource_id' in this.task) {
        // This is already an AsyncTask
        this.asyncTask = this.task as AsyncTask;
        this.taskId = this.task.id;
        this.loading = false;
        if (this.autoRefresh && (this.task.status === 'pending' || this.task.status === 'running')) {
          this.startPolling();
        }
      } else {
        // This is a simple Task object, load the full AsyncTask
        this.taskId = this.task.id;
        this.loadTask();
        if (this.autoRefresh) {
          this.startPolling();
        }
      }
    } else {
      this.error = 'No task ID provided';
      this.loading = false;
    }
  }

  ngOnDestroy(): void {
    this.stopPolling();
  }

  loadTask(): void {
    if (!this.taskId) {
      this.error = 'No task ID provided';
      this.loading = false;
      return;
    }

    this.loading = true;
    this.error = '';

    this.repositoryService.getTask(this.taskId).subscribe({
      next: (task) => {
        this.asyncTask = task;
        this.loading = false;
      },
      error: (err) => {
        this.error = 'Failed to load task: ' + (err.message || 'Unknown error');
        this.loading = false;
      }
    });
  }

  startPolling(): void {
    if (!this.taskId) {
      return;
    }

    // Stop any existing polling
    this.stopPolling();

    // Start new polling
    this.subscription = interval(this.refreshInterval).pipe(
      switchMap(() => this.repositoryService.getTask(this.taskId!)),
      takeWhile((task: AsyncTask) => task.status === 'pending' || task.status === 'running', true)
    ).subscribe({
      next: (task: AsyncTask) => {
        this.asyncTask = task;
        this.loading = false;

        // Stop polling if task is complete
        if (task.status !== 'pending' && task.status !== 'running') {
          this.stopPolling();
        }
      },
      error: (err: any) => {
        this.error = 'Failed to update task status: ' + (err.message || 'Unknown error');
        this.loading = false;
        this.stopPolling();
      }
    });
  }

  stopPolling(): void {
    if (this.subscription) {
      this.subscription.unsubscribe();
      this.subscription = null;
    }
  }

  // Helper methods for template
  getStatusClass(): string {
    if (!this.asyncTask) return '';

    switch (this.asyncTask.status) {
      case 'completed': return 'status-completed';
      case 'failed': return 'status-failed';
      case 'running': return 'status-running';
      case 'pending': return 'status-pending';
      default: return '';
    }
  }

  getStatusIcon(): string {
    if (!this.asyncTask) return '';

    switch (this.asyncTask.status) {
      case 'completed': return 'fa-check-circle';
      case 'failed': return 'fa-times-circle';
      case 'running': return 'fa-spinner fa-spin';
      case 'pending': return 'fa-clock';
      default: return 'fa-question-circle';
    }
  }

  getStatusIconName(): string {
    if (!this.asyncTask) return 'help';

    switch (this.asyncTask.status) {
      case 'completed': return 'check_circle';
      case 'failed': return 'error';
      case 'running': return 'refresh';
      case 'pending': return 'schedule';
      default: return 'help';
    }
  }

  formatDate(dateString: string | undefined): string {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleString();
  }

  getElapsedTime(): string {
    if (!this.asyncTask || !this.asyncTask.started_at) return 'N/A';

    const start = new Date(this.asyncTask.started_at).getTime();
    const end = this.asyncTask.completed_at
      ? new Date(this.asyncTask.completed_at).getTime()
      : new Date().getTime();

    const seconds = Math.floor((end - start) / 1000);

    if (seconds < 60) return `${seconds} seconds`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)} minutes`;
    return `${Math.floor(seconds / 3600)} hours`;
  }
}
