import { Component, OnInit, Input, Output, EventEmitter, OnChanges, SimpleChanges } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormArray, FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatIconModule } from '@angular/material/icon';
import { MatDatepickerModule } from '@angular/material/datepicker';
import { MatNativeDateModule } from '@angular/material/core';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import {MatChipsModule } from '@angular/material/chips';
import { COMMA, ENTER } from '@angular/cdk/keycodes';
import { CollectionFieldDefinition } from '../../services/repository.service';
import { MatCard, MatCardContent, MatCardHeader, MatCardSubtitle, MatCardTitle } from '@angular/material/card';
import { MatChipInputEvent } from '@angular/material/chips';

@Component({
  selector: 'app-front-matter-editor',
  templateUrl: './front-matter-editor.component.html',
  styleUrls: ['./front-matter-editor.component.scss'],
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatButtonModule,
    MatFormFieldModule,
    MatInputModule,
    MatIconModule,
    MatDatepickerModule,
    MatNativeDateModule,
    MatSlideToggleModule,
    MatCardContent,
    MatCardSubtitle,
    MatCardTitle,
    MatCard,
    MatCardHeader,
    MatChipsModule
  ]
})
export class FrontMatterEditorComponent implements OnInit, OnChanges {
  @Input() frontMatter: Record<string, any> = {};
  @Input() fields: CollectionFieldDefinition[] = [];
  @Output() frontMatterChange = new EventEmitter<Record<string, any>>();

  frontMatterForm!: FormGroup;
  readonly separatorKeysCodes = [ENTER, COMMA] as const;
  tags: string[] = [];

  constructor(private fb: FormBuilder) {}

  ngOnInit(): void {
    this.initForm();
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['frontMatter'] && !changes['frontMatter'].firstChange) {
      this.initForm();
    }
  }

  initForm(): void {
    this.tags = this.getTagsValue();
    this.frontMatterForm = this.fb.group({
      fields: this.fb.array([]),
      date: [this.getDateValue()],
      draft: [this.getDraftValue()]
    });

    const fieldsArray = this.frontMatterForm.get('fields') as FormArray;
    this.fields.forEach(field => {
      const defaultValue = this.frontMatter[field.name] !== undefined ? this.frontMatter[field.name] : field.default;
      const control = this.fb.control(defaultValue, this.getValidators(field));
      fieldsArray.push(control);
    });

    // Listen for changes to update front matter
    this.frontMatterForm.valueChanges.subscribe(() => {
      this.updateFrontMatter();
    });
  }

  addTag(event: MatChipInputEvent): void {
    const value = (event.value || '').trim();
    if (value) {
      this.tags.push(value);
      this.updateFrontMatter();
    }
    event.chipInput!.clear();
  }

  removeTag(tag: string): void {
    const index = this.tags.indexOf(tag);
    if (index >= 0) {
      this.tags.splice(index, 1);
      this.updateFrontMatter();
    }
  }

  updateFrontMatter(): void {
    if (!this.frontMatterForm.valid) return;

    const formValue = this.frontMatterForm.value;
    const updatedFrontMatter: Record<string, any> = { ...this.frontMatter };

    const fieldsArray = formValue.fields;
    this.fields.forEach((field, index) => {
      updatedFrontMatter[field.name] = fieldsArray[index];
    });

    if (formValue.date) {
      updatedFrontMatter['date'] = formValue.date;
    }
    if (formValue.draft !== undefined) {
      updatedFrontMatter['draft'] = formValue.draft;
    }
    updatedFrontMatter['tags'] = this.tags;

    this.frontMatterChange.emit(updatedFrontMatter);
  }

  getDateValue(): Date | null {
    const dateStr = this.frontMatter['date'];
    if (!dateStr) {
      return new Date();
    }
    const date = new Date(dateStr);
    return isNaN(date.getTime()) ? new Date() : date;
  }

  getDraftValue(): boolean {
    return this.frontMatter['draft'] === true;
  }

  getTagsValue(): string[] {
    return this.frontMatter['tags'] || [];
  }

  getValidators(field: CollectionFieldDefinition) {
    const validators = [];
    if (field.required) {
      validators.push(Validators.required);
    }
    return validators;
  }
}
