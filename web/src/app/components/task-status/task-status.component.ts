import { Component, Input, OnInit, OnDestroy } from '@angular/core';
import { AsyncTask, RepositoryService, Task } from '../../services/repository.service';
import { interval, Subscription } from 'rxjs';
import { switchMap, takeWhile } from 'rxjs/operators';
import { NgClass, TitleCasePipe } from '@angular/common';
import { MatButtonModule, MatIconButton } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatProgressBarModule } from '@angular/material/progress-bar';

@Component({
  selector: 'app-task-status',
  templateUrl: './task-status.component.html',
  imports: [
    NgClass,
    TitleCasePipe,
    MatButtonModule,
    MatIconButton,
    MatIconModule,
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
  detailsExpanded = false;
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
        this.setTask(this.task as AsyncTask);
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
        this.setTask(task);
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
        this.setTask(task);
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

  getSummaryMessage(): string {
    if (!this.asyncTask) return '';

    if (this.asyncTask.type === 'sync') {
      switch (this.asyncTask.status) {
        case 'completed': return 'Sync completed';
        case 'failed': return 'Sync failed';
        case 'running': return 'Syncing repository...';
        case 'pending': return 'Sync queued';
      }
    }

    switch (this.asyncTask.status) {
      case 'completed': return `${this.asyncTask.type} completed`;
      case 'failed': return `${this.asyncTask.type} failed`;
      case 'running': return `${this.asyncTask.type} running`;
      case 'pending': return `${this.asyncTask.type} pending`;
      default: return `${this.asyncTask.type} status`;
    }
  }

  getShortTaskId(): string {
    if (!this.asyncTask?.id) return 'N/A';
    if (this.asyncTask.id.length <= 16) return this.asyncTask.id;
    return `${this.asyncTask.id.slice(0, 8)}...${this.asyncTask.id.slice(-4)}`;
  }

  toggleDetails(): void {
    this.detailsExpanded = !this.detailsExpanded;
  }

  copyTaskId(): void {
    if (!this.asyncTask?.id || !navigator.clipboard) return;
    navigator.clipboard.writeText(this.asyncTask.id);
  }

  private setTask(task: AsyncTask): void {
    this.asyncTask = task;
    if (task.status === 'failed') {
      this.detailsExpanded = true;
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
