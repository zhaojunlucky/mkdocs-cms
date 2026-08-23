import { Component, OnInit, Input, Output, EventEmitter, OnChanges, SimpleChanges, ChangeDetectionStrategy, OnDestroy } from '@angular/core';

import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatIconModule } from '@angular/material/icon';
import { MatDatepickerModule } from '@angular/material/datepicker';
import { MatNativeDateModule } from '@angular/material/core';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { MatChipsModule } from '@angular/material/chips';
import { COMMA, ENTER } from '@angular/cdk/keycodes';
import { CollectionFieldDefinition, SeoIssue, SeoPage, SeoReport } from '../../services/repository.service';
import { MatChipInputEvent } from '@angular/material/chips';
import { Subscription } from 'rxjs';
import { MarkdownSeoUtils } from '../../shared/utils/markdown-seo.utils';

@Component({
  selector: 'app-front-matter-editor',
  templateUrl: './front-matter-editor.component.html',
  styleUrls: ['./front-matter-editor.component.scss'],
  standalone: true,
  changeDetection: ChangeDetectionStrategy.Eager,
  imports: [
    ReactiveFormsModule,
    MatButtonModule,
    MatFormFieldModule,
    MatInputModule,
    MatIconModule,
    MatDatepickerModule,
    MatNativeDateModule,
    MatSlideToggleModule,
    MatChipsModule
  ]
})
export class FrontMatterEditorComponent implements OnInit, OnChanges, OnDestroy {
  @Input() frontMatter: Record<string, any> = {};
  @Input() fields: CollectionFieldDefinition[] = [];
  @Input() markdownContent = '';
  @Input() seoReport: SeoReport | null = null;
  @Input() collectionPath = '';
  @Input() filePath = '';
  @Input() disabled = false
  @Output() frontMatterChange = new EventEmitter<Record<string, any>>();
  @Output() frontMatterInit = new EventEmitter<Record<string, any>>();
  @Output() insertMarkdown = new EventEmitter<string>();

  frontMatterForm!: FormGroup;
  readonly separatorKeysCodes = [ENTER, COMMA] as const;
  listValues: Record<string, string[]> = {}
  seoFields: CollectionFieldDefinition[] = [];
  metadataFields: CollectionFieldDefinition[] = [];

  private formSubscription?: Subscription;
  private autoManagedFields = new Set<string>();

  constructor(private fb: FormBuilder) {}

  ngOnInit(): void {
    this.initForm();
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (!this.frontMatterForm) {
      return;
    }

    if (changes['fields']) {
      this.initForm();
      return;
    }

    if (changes['frontMatter'] && this.hasExternalFrontMatterChange()) {
      this.initForm();
      return;
    }

    if (changes['markdownContent']) {
      this.refreshAutoManagedFields();
    }
  }

  ngOnDestroy(): void {
    this.formSubscription?.unsubscribe();
  }

  initForm(): void {
    this.formSubscription?.unsubscribe();
    this.listValues = {};
    this.autoManagedFields.clear();
    this.frontMatterForm = this.fb.group({});

    const fields = this.orderFields(this.fields.filter(f => f.name !== 'body'));
    this.seoFields = fields.filter(field => this.isSeoField(field));
    this.metadataFields = fields.filter(field => !this.isSeoField(field));
    let hasGeneratedFrontMatter = false;

    fields.forEach(field => {
      this.frontMatterForm.addControl(field.name, this.fb.control('', field.required ? [Validators.required] : []));
      let defaultValue = this.frontMatter[field.name] !== undefined && this.frontMatter[field.name] !== null ? this.frontMatter[field.name] : field.default;
      if ((defaultValue === undefined || defaultValue === null || defaultValue === '') && field.auto_from) {
        defaultValue = this.generateValue(field);
        if (defaultValue) {
          this.autoManagedFields.add(field.name);
          hasGeneratedFrontMatter = true;
        }
      }

      switch (field.type) {
        case 'date':
          defaultValue = defaultValue ? new Date(defaultValue) : (field.required ? new Date() : null);
          break
        case 'string':
        case 'text': {
          if (field.list) {
            defaultValue = defaultValue instanceof Array ? defaultValue : []
            this.listValues[field.name] = defaultValue
          } else {
            defaultValue = defaultValue ?? ''
          }
          break;
        }
        case 'boolean': defaultValue = defaultValue == true || defaultValue == 'true'; break
        case 'number': defaultValue = defaultValue === undefined || defaultValue === null || defaultValue === '' ? null : Number(defaultValue); break
        default: defaultValue = defaultValue ?? ''; break;
      }

      this.frontMatterForm.patchValue({
        [field.name]: defaultValue
      }, { emitEvent: false })
    });
    this.updateFrontMatter(true);
    if (hasGeneratedFrontMatter) {
      this.updateFrontMatter();
    }
    this.formSubscription = this.frontMatterForm.valueChanges.subscribe(() => {
      this.updateFrontMatter();
    });
  }

  private orderFields(fields: CollectionFieldDefinition[]): CollectionFieldDefinition[] {
    const orderedFields = [...fields];
    const draftIndex = orderedFields.findIndex(field => field.name === 'draft');
    const tagsIndex = orderedFields.findIndex(field => field.name === 'tags');

    if (draftIndex > tagsIndex && tagsIndex >= 0) {
      const [draftField] = orderedFields.splice(draftIndex, 1);
      orderedFields.splice(tagsIndex, 0, draftField);
    }

    return orderedFields;
  }

  private isSeoField(field: CollectionFieldDefinition): boolean {
    return !!field.seo || field.name === 'title' || field.name === 'description';
  }

  private hasExternalFrontMatterChange(): boolean {
    if (!this.frontMatterForm) {
      return false;
    }

    return this.fields
      .filter(field => field.name !== 'body')
      .some(field => !this.areFieldValuesEqual(this.frontMatterForm.get(field.name)?.value, this.frontMatter[field.name], field));
  }

  private areFieldValuesEqual(formValue: any, inputValue: any, field: CollectionFieldDefinition): boolean {
    if (field.type === 'date') {
      return this.dateValue(formValue) === this.dateValue(inputValue);
    }

    if (Array.isArray(formValue) || Array.isArray(inputValue)) {
      return JSON.stringify(formValue || []) === JSON.stringify(inputValue || []);
    }

    return (formValue ?? '') === (inputValue ?? field.default ?? '');
  }

  private dateValue(value: any): string {
    if (!value) {
      return '';
    }
    const date = value instanceof Date ? value : new Date(value);
    return Number.isNaN(date.getTime()) ? '' : date.toISOString().slice(0, 10);
  }

  addTag(event: MatChipInputEvent, name: string): void {
    const value = (event.value || '').trim();
    if (value && this.listValues[name].findIndex(v => v == value) < 0) {
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

  markManualEdit(name: string): void {
    this.autoManagedFields.delete(name);
  }

  autoFill(field: CollectionFieldDefinition): void {
    const value = this.generateValue(field);
    if (!value) {
      return;
    }

    this.autoManagedFields.add(field.name);
    this.frontMatterForm.patchValue({ [field.name]: value }, { emitEvent: false });
    this.updateFrontMatter();
  }

  descriptionLength(fieldName: string): number {
    return String(this.frontMatterForm.get(fieldName)?.value || '').trim().length;
  }

  get currentSourcePath(): string {
    return this.normalizePath([this.collectionPath, this.filePath].filter(Boolean).join('/'));
  }

  get currentPage(): SeoPage | null {
    if (!this.seoReport || !this.currentSourcePath) {
      return null;
    }
    return this.seoReport.pages.find(page => page.sourcePath === this.currentSourcePath) || null;
  }

  get canonicalUrl(): string {
    if (this.currentPage?.canonicalUrl) {
      return this.currentPage.canonicalUrl;
    }
    if (!this.seoReport || !this.currentSourcePath) {
      return '';
    }
    return this.joinUrl(this.seoReport.site.siteUrl, this.buildUrlPath(this.currentSourcePath));
  }

  get searchPreviewTitle(): string {
    const title = String(this.frontMatterForm.get('title')?.value || '').trim()
      || MarkdownSeoUtils.extractFirstH1(this.markdownContent)
      || 'Untitled';
    const siteName = this.seoReport?.site.siteName;
    return siteName ? `${title} - ${siteName}` : title;
  }

  get searchPreviewDescription(): string {
    return String(this.frontMatterForm.get('description')?.value || '').trim()
      || 'Add a description to control the search result summary.';
  }

  get currentTags(): string[] {
    const tags = this.frontMatterForm.get('tags')?.value;
    return Array.isArray(tags) ? tags : [];
  }

  get relatedPages(): SeoPage[] {
    if (!this.seoReport || this.currentTags.length === 0) {
      return [];
    }
    const tagSet = new Set(this.currentTags.map(tag => String(tag).toLowerCase()));
    return this.seoReport.pages
      .filter(page => page.sourcePath !== this.currentSourcePath && !page.draft)
      .map(page => ({
        page,
        score: (page.tags || []).filter(tag => tagSet.has(String(tag).toLowerCase())).length
      }))
      .filter(item => item.score > 0)
      .sort((a, b) => b.score - a.score || a.page.sourcePath.localeCompare(b.page.sourcePath))
      .slice(0, 5)
      .map(item => item.page);
  }

  get siteHealthIssues(): SeoIssue[] {
    return (this.seoReport?.issues || [])
      .filter(issue => !issue.path)
      .filter(issue => issue.type === 'missing_site_url'
        || issue.type === 'missing_site_description'
        || issue.type === 'robots_blocks_all'
        || issue.type === 'sitemap_not_built')
      .slice(0, 5);
  }

  fieldWarnings(field: CollectionFieldDefinition): string[] {
    const value = String(this.frontMatterForm.get(field.name)?.value || '').trim();
    const warnings: string[] = [];

    if (field.name === 'title') {
      const h1 = MarkdownSeoUtils.extractFirstH1(this.markdownContent);
      const h1Count = MarkdownSeoUtils.countH1(this.markdownContent);
      if (!value) {
        warnings.push('Title is missing.');
      }
      if (!h1) {
        warnings.push('No H1 found.');
      }
      if (h1Count > 1) {
        warnings.push('More than one H1 found.');
      }
      if (value && h1 && !MarkdownSeoUtils.isSimilarTitle(value, h1)) {
        warnings.push('Title differs from H1.');
      }
    }

    if (field.name === 'description') {
      const min = field.recommended_min || 80;
      const max = field.recommended_max || 160;
      if (!value) {
        warnings.push('Description is missing.');
      } else if (value.length < min) {
        warnings.push(`Description is shorter than ${min} characters.`);
      } else if (value.length > max) {
        warnings.push(`Description is longer than ${max} characters.`);
      }
      if (/[[\]()`*_#]/.test(value)) {
        warnings.push('Description contains markdown syntax.');
      }
      if (value.startsWith('<!--')) {
        warnings.push('Description starts with an HTML comment.');
      }
      if (value && value === String(this.frontMatterForm.get('title')?.value || '').trim()) {
        warnings.push('Description duplicates the title.');
      }
    }

    warnings.push(...this.duplicateWarnings(field, value));

    return warnings;
  }

  pageIssues(): SeoIssue[] {
    return (this.seoReport?.issues || [])
      .filter(issue => issue.path === this.currentSourcePath)
      .slice(0, 5);
  }

  insertRelatedLink(page: SeoPage): void {
    const label = page.title || page.h1 || page.sourcePath;
    this.insertMarkdown.emit(`[${label}](${page.urlPath || page.canonicalUrl})`);
  }

  updateFrontMatter(init: boolean = false): void {
    if (!this.frontMatterForm.valid) return;

    const formValue = this.frontMatterForm.value;
    const updatedFrontMatter: Record<string, any> = { ...this.frontMatter };
    this.fields.filter(field => field.name !== 'body').forEach(field => {
      const value = formValue[field.name];
      if (this.shouldOmitEmptyValue(field, value)) {
        delete updatedFrontMatter[field.name];
        return;
      }
      updatedFrontMatter[field.name] = value
    });
    if (init) {
      this.frontMatterInit.emit(updatedFrontMatter);
    } else {
      this.frontMatterChange.emit(updatedFrontMatter);
    }
  }

  private refreshAutoManagedFields(): void {
    if (!this.frontMatterForm) {
      return;
    }

    let updated = false;
    for (const field of this.seoFields) {
      if (!field.auto_from) {
        continue;
      }
      const control = this.frontMatterForm.get(field.name);
      const currentValue = String(control?.value || '').trim();
      const shouldAutoFill = this.autoManagedFields.has(field.name) || currentValue === '';
      if (!shouldAutoFill) {
        continue;
      }

      const value = this.generateValue(field);
      if (value && control?.value !== value) {
        this.autoManagedFields.add(field.name);
        this.frontMatterForm.patchValue({ [field.name]: value }, { emitEvent: false });
        updated = true;
      }
    }

    if (updated) {
      this.updateFrontMatter();
    }
  }

  private generateValue(field: CollectionFieldDefinition): string {
    if (field.auto_from === 'h1' || field.name === 'title') {
      return MarkdownSeoUtils.extractFirstH1(this.markdownContent);
    }

    if (field.auto_from === 'excerpt' || field.auto_from === 'content' || field.name === 'description') {
      return MarkdownSeoUtils.generateDescription(this.markdownContent, field.recommended_max || 160);
    }

    return '';
  }

  private shouldOmitEmptyValue(field: CollectionFieldDefinition, value: any): boolean {
    if (!field.seo || field.required || field.type === 'boolean') {
      return false;
    }
    if (value === undefined || value === null || value === '') {
      return true;
    }
    return Array.isArray(value) && value.length === 0;
  }

  private duplicateWarnings(field: CollectionFieldDefinition, value: string): string[] {
    if (!this.seoReport || !value || (field.name !== 'title' && field.name !== 'description')) {
      return [];
    }
    const normalizedValue = value.trim().toLowerCase();
    const duplicate = this.seoReport.pages.find(page => {
      if (page.sourcePath === this.currentSourcePath || page.draft) {
        return false;
      }
      const pageValue = field.name === 'title' ? page.title : page.description;
      return pageValue.trim().toLowerCase() === normalizedValue;
    });
    return duplicate ? [`${field.label} duplicates ${duplicate.sourcePath}.`] : [];
  }

  private buildUrlPath(sourcePath: string): string {
    if (!this.seoReport) {
      return '';
    }

    const site = this.seoReport.site;
    const postPrefix = this.normalizePath(site.postDir) + '/';
    if (sourcePath.startsWith(postPrefix)) {
      const fileName = sourcePath.split('/').pop()?.replace(/\.md$/, '') || '';
      const match = fileName.match(/^(\d{4}-\d{2}-\d{2})-(.+)$/);
      const date = this.dateValue(this.frontMatterForm.get('date')?.value) || match?.[1] || '';
      const slug = this.slugify(match?.[2] || fileName);
      const postPath = (site.postUrlFormat || '{date}/{slug}')
        .replaceAll('{date}', date)
        .replaceAll('{slug}', slug);
      return `/${this.trimSlashes(site.blogDir || 'blog')}/${this.trimSlashes(postPath)}/`;
    }

    const docsPrefix = this.normalizePath(site.docsDir || 'docs') + '/';
    let path = sourcePath.startsWith(docsPrefix) ? sourcePath.slice(docsPrefix.length) : sourcePath;
    path = path.replace(/\.md$/, '').replace(/\/index$/, '');
    return site.useDirectoryUrls ? `/${this.trimSlashes(path)}/` : `/${this.trimSlashes(path)}.html`;
  }

  private joinUrl(siteUrl: string, path: string): string {
    if (!siteUrl) {
      return path;
    }
    return `${siteUrl.replace(/\/$/, '')}${path}`;
  }

  private normalizePath(path: string): string {
    return path.replace(/\\/g, '/').replace(/\/+/g, '/').replace(/^\/|\/$/g, '');
  }

  private trimSlashes(value: string): string {
    return value.replace(/^\/|\/$/g, '');
  }

  private slugify(value: string): string {
    return value
      .toLowerCase()
      .trim()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-|-$/g, '');
  }
}
