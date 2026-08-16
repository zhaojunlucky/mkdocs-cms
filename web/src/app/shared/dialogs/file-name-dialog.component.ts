import { ChangeDetectionStrategy, Component, Inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatIconModule } from '@angular/material/icon';
import { FileNameUtils } from '../utils/file-name.utils';

export interface FileNameDialogData {
  title: string;
  prefix: string;
  suffix?: string;
  extension?: string;
  markdownContent?: string;
  confirmLabel?: string;
}

@Component({
  selector: 'app-file-name-dialog',
  standalone: true,
  imports: [
    FormsModule,
    MatButtonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule
  ],
  changeDetection: ChangeDetectionStrategy.Eager,
  template: `
    <form (ngSubmit)="confirm()">
      <h2 mat-dialog-title>{{ data.title }}</h2>
      <mat-dialog-content>
        <div class="file-name-form">
          <div class="readonly-field">
            <span>Prefix</span>
            <strong>{{ data.prefix || 'None' }}</strong>
          </div>

          <mat-form-field appearance="outline">
            <mat-label>Suffix</mat-label>
            <input matInput name="suffix" [(ngModel)]="suffix" placeholder="my-post-title" cdkFocusInitial>
            @if (error) {
              <mat-error>{{ error }}</mat-error>
            }
          </mat-form-field>

          <div class="readonly-field">
            <span>File name</span>
            <strong>{{ fileNamePreview }}</strong>
          </div>
        </div>
      </mat-dialog-content>
      <mat-dialog-actions align="end">
        <button mat-button type="button" (click)="autoFromH1()">
          <mat-icon>auto_fix_high</mat-icon>
          <span>Auto from H1</span>
        </button>
        <span class="actions-spacer"></span>
        <button mat-button type="button" [mat-dialog-close]="null">Cancel</button>
        <button mat-flat-button color="primary" type="submit" [disabled]="!isValidSuffix">
          {{ data.confirmLabel || 'Save' }}
        </button>
      </mat-dialog-actions>
    </form>
  `,
  styles: [`
    .file-name-form {
      display: flex;
      flex-direction: column;
      gap: 12px;
      min-width: min(420px, 70vw);
      padding-top: 8px;
    }

    mat-form-field {
      width: 100%;
    }

    .readonly-field {
      display: flex;
      flex-direction: column;
      gap: 4px;
      padding: 8px 0;
    }

    .readonly-field span {
      color: var(--mat-sys-on-surface-variant);
      font: var(--mat-sys-label-medium);
    }

    .readonly-field strong {
      overflow-wrap: anywhere;
      font: var(--mat-sys-body-large);
      font-weight: 500;
    }

    .actions-spacer {
      flex: 1 1 auto;
    }
  `]
})
export class FileNameDialogComponent {
  suffix = '';
  error = '';

  constructor(
    private dialogRef: MatDialogRef<FileNameDialogComponent, string | null>,
    @Inject(MAT_DIALOG_DATA) public data: FileNameDialogData
  ) {
    this.suffix = data.suffix || '';
  }

  get extension(): string {
    return this.data.extension || '.md';
  }

  get fileNamePreview(): string {
    return FileNameUtils.buildFileName(this.data.prefix, this.suffix.trim() || '{title}', this.extension);
  }

  get isValidSuffix(): boolean {
    return FileNameUtils.isValidSuffix(this.suffix);
  }

  autoFromH1(): void {
    const suffix = FileNameUtils.slugifyTitle(FileNameUtils.extractFirstH1(this.data.markdownContent || ''));
    if (!suffix) {
      this.error = 'No H1 heading found. Enter a suffix manually.';
      return;
    }

    this.suffix = suffix;
    this.error = '';
  }

  confirm(): void {
    const suffix = this.suffix.trim();
    if (!FileNameUtils.isValidSuffix(suffix)) {
      this.error = 'Enter a valid file name suffix.';
      return;
    }

    this.dialogRef.close(suffix);
  }
}
