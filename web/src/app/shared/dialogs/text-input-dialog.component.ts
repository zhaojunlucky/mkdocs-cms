import { ChangeDetectionStrategy, Component, Inject, OnDestroy } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { Observable, Subscription } from 'rxjs';

export interface TextInputDialogData {
  title: string;
  label: string;
  value?: string;
  placeholder?: string;
  hint?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  validator?: (value: string) => string | null;
  asyncValidator?: (value: string) => Observable<string | null>;
}

@Component({
  selector: 'app-text-input-dialog',
  standalone: true,
  imports: [
    FormsModule,
    MatButtonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule
  ],
  changeDetection: ChangeDetectionStrategy.Eager,
  template: `
    <form (ngSubmit)="confirm()">
      <h2 mat-dialog-title>{{ data.title }}</h2>
      <mat-dialog-content>
        <div class="field-wrapper">
          <mat-form-field appearance="outline">
            <mat-label>{{ data.label }}</mat-label>
            <input matInput name="value" [(ngModel)]="value" [placeholder]="data.placeholder || ''" (ngModelChange)="validate()" cdkFocusInitial>
            @if (validationPending) {
              <mat-hint>Checking...</mat-hint>
            }
            @if (!validationPending && data.hint) {
              <mat-hint>{{ data.hint }}</mat-hint>
            }
          </mat-form-field>
          @if (validationMessage) {
            <p class="validation-message">{{ validationMessage }}</p>
          }
        </div>
      </mat-dialog-content>
      <mat-dialog-actions align="end">
        <button mat-button type="button" [mat-dialog-close]="null">{{ data.cancelLabel || 'Cancel' }}</button>
        <button mat-flat-button color="primary" type="submit" [disabled]="!canConfirm">
          {{ data.confirmLabel || 'Save' }}
        </button>
      </mat-dialog-actions>
    </form>
  `,
  styles: [`
    .field-wrapper {
      padding-top: 8px;
    }

    mat-form-field {
      width: 100%;
      min-width: min(360px, 70vw);
    }

    .validation-message {
      margin: -14px 0 0;
      color: var(--mat-sys-error);
      font: var(--mat-sys-body-small);
    }
  `]
})
export class TextInputDialogComponent implements OnDestroy {
  value: string;
  validationMessage = '';
  validationPending = false;
  private validationSubscription?: Subscription;
  private validationRun = 0;

  constructor(
    private dialogRef: MatDialogRef<TextInputDialogComponent, string | null>,
    @Inject(MAT_DIALOG_DATA) public data: TextInputDialogData
  ) {
    this.value = data.value || '';
    this.validate();
  }

  get canConfirm(): boolean {
    return !!this.value.trim() && !this.validationMessage && !this.validationPending;
  }

  validate(): void {
    const trimmedValue = this.value.trim();
    const run = ++this.validationRun;

    this.validationSubscription?.unsubscribe();
    this.validationPending = false;
    this.validationMessage = '';

    if (!trimmedValue) {
      return;
    }

    const validationMessage = this.data.validator?.(trimmedValue);
    if (validationMessage) {
      this.validationMessage = validationMessage;
      return;
    }

    if (!this.data.asyncValidator) {
      return;
    }

    this.validationPending = true;
    this.validationSubscription = this.data.asyncValidator(trimmedValue).subscribe({
      next: (asyncMessage) => {
        if (run !== this.validationRun) return;
        this.validationPending = false;
        this.validationMessage = asyncMessage || '';
      },
      error: () => {
        if (run !== this.validationRun) return;
        this.validationPending = false;
        this.validationMessage = 'Could not validate this value.';
      }
    });
  }

  confirm(): void {
    const trimmedValue = this.value.trim();
    if (this.canConfirm) {
      this.dialogRef.close(trimmedValue);
    }
  }

  ngOnDestroy(): void {
    this.validationSubscription?.unsubscribe();
  }
}
