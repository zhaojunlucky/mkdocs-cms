import { Component, OnInit, Input, Output, EventEmitter, OnChanges, SimpleChanges } from '@angular/core';

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
export class FrontMatterEditorComponent implements OnInit {
  @Input() frontMatter: Record<string, any> = {};
  @Input() fields: CollectionFieldDefinition[] = [];
  @Output() frontMatterChange = new EventEmitter<Record<string, any>>();
  @Output() frontMatterInit = new EventEmitter<Record<string, any>>();
  @Input() disabled = false

  frontMatterForm!: FormGroup;
  readonly separatorKeysCodes = [ENTER, COMMA] as const;
  listValues: Record<string, string[]> = {}

  constructor(private fb: FormBuilder) {}

  ngOnInit(): void {
    this.initForm();
  }

  initForm(): void {
    this.frontMatterForm = this.fb.group({
    });
    this.fields = this.fields.filter(f=>f.name !== 'body')

    this.fields.forEach(field => {
      this.frontMatterForm.addControl(field.name, this.fb.control('', field.required?[Validators.required]:[]));
      let defaultValue = this.frontMatter[field.name] !== undefined && this.frontMatter[field.name] !== null ? this.frontMatter[field.name] : field.default;
      switch (field.type) {
        case 'date': defaultValue = defaultValue instanceof Date? defaultValue: new Date(); break
        case 'string':  {
          if (field.list) {
            defaultValue = defaultValue instanceof Array? defaultValue: []
            this.listValues[field.name] = defaultValue
          } else {
            defaultValue = ''
          }
          break;
        }
        case 'boolean': defaultValue = defaultValue == true || defaultValue == 'true'; break
        default: break;
      }

      this.frontMatterForm.patchValue({
        [field.name]: defaultValue
      })
    });
    this.updateFrontMatter(true);
    // Listen for changes to update front matter
    this.frontMatterForm.valueChanges.subscribe(() => {
      this.updateFrontMatter();
    });
  }

  addTag(event: MatChipInputEvent, name: string): void {
    const value = (event.value || '').trim();
    if (value && this.listValues[name].findIndex(v=> v == value) < 0) {
      this.listValues[name].push(value);
      this.updateFrontMatter();
    }
    event.chipInput!.clear();
  }

  removeTag(tag: string, name: string): void {
    const index = this.listValues[name].indexOf(tag);
    if (index >= 0) {
      this.listValues[name].splice(index, 1);
      this.updateFrontMatter();
    }
  }

  updateFrontMatter(init:boolean = false): void {
    if (!this.frontMatterForm.valid) return;

    const formValue = this.frontMatterForm.value;
    const updatedFrontMatter: Record<string, any> = { ...this.frontMatter };
    Object.entries(formValue).forEach(([key, value]) => {
      updatedFrontMatter[key] = value
    });
    if (init) {
      this.frontMatterInit.emit(updatedFrontMatter);
    } else {
      this.frontMatterChange.emit(updatedFrontMatter);
    }
  }
}
