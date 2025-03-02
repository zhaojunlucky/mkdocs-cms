import { Component, EventEmitter, Input, OnChanges, OnInit, Output, SimpleChanges } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormArray, FormBuilder, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatIconModule } from '@angular/material/icon';
import { MatCardModule } from '@angular/material/card';
import { MatDividerModule } from '@angular/material/divider';

@Component({
  selector: 'app-front-matter-editor',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatButtonModule,
    MatFormFieldModule,
    MatInputModule,
    MatIconModule,
    MatCardModule,
    MatDividerModule
  ],
  templateUrl: './front-matter-editor.component.html',
  styleUrls: ['./front-matter-editor.component.scss']
})
export class FrontMatterEditorComponent implements OnInit, OnChanges {
  @Input() frontMatter: Record<string, any> = {};
  @Output() frontMatterChange = new EventEmitter<Record<string, any>>();
  
  frontMatterForm!: FormGroup;
  
  constructor(private fb: FormBuilder) {}
  
  ngOnInit(): void {
    this.initForm();
  }
  
  ngOnChanges(changes: SimpleChanges): void {
    if (changes['frontMatter'] && !this.isEqual(changes['frontMatter'].previousValue, changes['frontMatter'].currentValue)) {
      this.initForm();
    }
  }
  
  private initForm(): void {
    this.frontMatterForm = this.fb.group({
      fields: this.fb.array([])
    });
    
    const fieldsArray = this.frontMatterForm.get('fields') as FormArray;
    
    // Clear existing fields
    while (fieldsArray.length > 0) {
      fieldsArray.removeAt(0);
    }
    
    // Add fields from frontMatter
    Object.entries(this.frontMatter).forEach(([key, value]) => {
      fieldsArray.push(
        this.fb.group({
          key: [key],
          value: [value]
        })
      );
    });
    
    // Add an empty field if there are no fields
    if (fieldsArray.length === 0) {
      this.addField();
    }
  }
  
  get fields(): FormArray {
    return this.frontMatterForm.get('fields') as FormArray;
  }
  
  addField(): void {
    this.fields.push(
      this.fb.group({
        key: [''],
        value: ['']
      })
    );
  }
  
  removeField(index: number): void {
    this.fields.removeAt(index);
    this.updateFrontMatter();
  }
  
  updateFrontMatter(): void {
    const updatedFrontMatter: Record<string, any> = {};
    
    this.fields.controls.forEach(control => {
      const key = control.get('key')?.value;
      const value = control.get('value')?.value;
      
      if (key && key.trim() !== '') {
        updatedFrontMatter[key] = value;
      }
    });
    
    this.frontMatterChange.emit(updatedFrontMatter);
  }
  
  private isEqual(obj1: any, obj2: any): boolean {
    if (obj1 === obj2) return true;
    if (!obj1 || !obj2) return false;
    
    const keys1 = Object.keys(obj1);
    const keys2 = Object.keys(obj2);
    
    if (keys1.length !== keys2.length) return false;
    
    for (const key of keys1) {
      if (obj1[key] !== obj2[key]) return false;
    }
    
    return true;
  }
}
