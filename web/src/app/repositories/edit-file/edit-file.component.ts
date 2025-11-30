import {Component, HostListener, NgZone, OnInit} from '@angular/core';

import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule, MatIconButton} from '@angular/material/button';
import {VditorEditorComponent} from '../../components/vditor-editor/vditor-editor.component';
import {Collection, RepositoryService} from '../../services/repository.service';
import {CollectionService} from '../../services/collection.service';
import * as yaml from 'js-yaml';
import {FrontMatterEditorComponent} from '../../markdown/front-matter-editor/front-matter-editor.component';
import {MatIcon} from '@angular/material/icon';
import {MatTooltip} from '@angular/material/tooltip';
import {ArrayResponse} from '../../shared/core/response';
import {StrUtils} from '../../shared/utils/str.utils';
import {CanComponentDeactivate} from '../../shared/guard/can-deactivate-form.guard';
import {Observable, of} from 'rxjs';
import {PageTitleService} from '../../services/page.title.service';
import {VditorUploadService} from '../../services/vditor.upload.service';
import {HttpHeaders} from '@angular/common/http';
import {MatSnackBar} from '@angular/material/snack-bar';

interface PathSegment {
  name: string;
  path: string;
}

@Component({
  selector: 'app-edit-file',
  standalone: true,
  imports: [
    FormsModule,
    RouterLink,
    MatProgressSpinnerModule,
    MatCardModule,
    MatButtonModule,
    FrontMatterEditorComponent,
    VditorEditorComponent,
    MatIcon,
    MatIconButton,
    MatTooltip
  ],
  templateUrl: './edit-file.component.html',
  styleUrls: ['./edit-file.component.scss']
})
export class EditFileComponent implements OnInit, CanComponentDeactivate {
  repositoryId: string = '';
  collectionName: string = '';
  filePath: string = '';
  fileName: string = '';

  collection: Collection | null | undefined = null;

  isLoading: boolean = true;
  error: string = '';
  fileError: string = '';

  // Editor state
  _markdownContent: string = '';
  frontMatter: Record<string, any> = {};
  savingFile: boolean = false;

  // Path navigation
  pathSegments: PathSegment[] = [];
  editorRendered = false;
  _changed = false;

  // Editor options
  editorOptions = {
    theme: 'classic',
    language: 'markdown',
    lang: 'en_US',
    icon: 'material',
    mode: 'wysiwyg',
    tab: '    ',
    counter: {
      enable: true,
    },
    cache: {
      enable: false,
    },
    toolbar: [
      'emoji',
      'headings',
      'bold',
      'italic',
      'strike',
      'link',
      '|',
      'list',
      'ordered-list',
      'check',
      'outdent',
      'indent',
      '|',
      'quote',
      'line',
      'code',
      'inline-code',
      'insert-before',
      'insert-after',
      '|',
      'upload',
      'record',
      'table',
      '|',
      'undo',
      'redo',
      '|',
      'edit-mode',

      {
        name: 'more',
        toolbar: [
          'code-theme',
          'content-theme',
          'export',
          'outline',
          'preview',
          'devtools',
          'info',
          'help',
        ],
      },
    ],
    preview: {
      hljs: {
        lineNumber: true,
      },
      markdown: {
        toc: true
      },
      actions: [
        "desktop"
      ]
    },
    height: this.calculateEditorHeight(),
  };

  editor: any = null;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private collectionService: CollectionService,
    private repositoryService: RepositoryService,
    private zone: NgZone,
    private pageTitleService: PageTitleService,
    private vditorUploadService: VditorUploadService,
    private snackBar: MatSnackBar

  ) {
    this.editorOptions = {...this.editorOptions, ...this.vditorUploadService.getVditorOptions()};
  }

  get changed(): boolean {
    return this._changed;
  }

  set changed(value: boolean) {
    this._changed = value;
    let title = this.pageTitleService.title;
    if (this._changed && !title.startsWith('*')) {
      title = '* ' + title;
    } else if (!this._changed && title.startsWith('*')) {
      title = title.substring(1).trimStart()
    }
    this.pageTitleService.title = title;
  }

  @HostListener('window:beforeunload', ['$event'])
  unloadNotification($event: any): void {
    if (this.changed) {
      $event.returnValue = true;
    }
  }

  @HostListener('document:visibilitychange', ['$event'])
  onVisibilityChange(event: Event): void {
    console.log('Visibility changed. Document hidden:', document.hidden, 'Has changes:', this.changed);

    if (document.hidden && this.changed) {
      console.log('Attempting to show notification. Permission:', Notification.permission);

      if (Notification.permission === 'granted') {
        const notification = new Notification('Unsaved Changes!', {
          body: "You have unsaved changes in your file. Please save to avoid losing your work!",
          icon: '/favicon.ico',
          requireInteraction: true, // Keep notification visible until user interacts
          tag: 'unsaved-changes' // Prevent duplicate notifications
        });

        notification.onclick = () => {
          window.focus(); // Bring the tab back to focus
          notification.close();
        };

        console.log('Notification created successfully');
      } else {
        console.log('Notification permission not granted, showing alert');
        alert("Please enable notification permission!!! You have unsaved changes, please save to avoid losing your work!!!");
      }
    }
  }


  @HostListener('window:resize')
  onWindowResize() {
    // console.log("window resized");
    this.updateEditorHeight();
  }

  @HostListener('document:keydown', ['$event'])
  handleKeyDown(event: KeyboardEvent) {
    if ((event.ctrlKey || event.metaKey) && event.key === 's') {
      event.preventDefault(); // Prevent browser's default Save action
      if (this.changed) {
        this.saveFile();

      }
    }
  }

  get markdownContent(): string {
    return this._markdownContent;
  }

  set markdownContent(value: string) {
    this._markdownContent = value;
    this.changed = true;
  }

  ngOnInit(): void {
    this.pageTitleService.title = 'Edit File';
    if (this.route.parent) {
      this.repositoryId = this.route.parent.snapshot.paramMap.get('id') || '';
    }
    this.route.paramMap.subscribe(params => {
      this.collectionName = params.get('collectionName') || '';

      this.route.queryParamMap.subscribe(queryParams => {
        this.filePath = queryParams.get('path') || '';
        if (this.repositoryId && this.collectionName && this.filePath) {
          this.fileName = this.filePath.split('/').pop() || '';
          this.pageTitleService.title = `Edit File - ${this.collectionName} - ${this.fileName}`
          this.setupPathSegments();
          this.loadData();
        } else {
          this.error = 'Invalid repository ID or collection name';
          this.isLoading = false;
        }

      });


    });
  }

  loadData(): void {
    this.isLoading = true;
    this.error = '';
    if (!this.repositoryId || !this.collectionName || !this.filePath) {
      this.isLoading = false;
      return;
    }
    this.loadCollection();
    this.loadFileContent();
  }

  loadCollection(): void {
    this.repositoryService.getRepositoryCollections(this.repositoryId).subscribe({
      next: (collections: ArrayResponse<Collection>) => {
        this.collection = collections.entries.find(c => c.name === this.collectionName);
        if (this.collection) {
          this.loadFileContent()
        } else {
          this.error = `Failed to load collection: ${this.collectionName}`;
          this.isLoading = false;
        }
      },
      error: (err: any) => {
        this.error = `Failed to load collection: ${err.message || 'Unknown error'}`;
        this.isLoading = false;
      }
    });
  }

  loadFileContent(): void {
    this.isLoading = true;

    this.collectionService.getFileContent(this.repositoryId, this.collectionName, this.filePath).subscribe({
      next: (fileContent) => {
        // Parse the file content to separate front matter and markdown
        this.parseFileContent(fileContent);
        this.isLoading = false;

      },
      error: (err: any) => {
        this.error = `Failed to load file content: ${err.message || 'Unknown error'}`;
        this.isLoading = false;
      }
    });
  }

  parseFileContent(content: string): void {
    // Check if the file has front matter (starts with ---)
    if (content.startsWith('---')) {
      const endOfFrontMatter = content.indexOf('---', 3);
      if (endOfFrontMatter !== -1) {
        const frontMatterText = content.substring(3, endOfFrontMatter).trim();
        try {
          this.frontMatter = yaml.load(frontMatterText) as Record<string, any>;
          this._markdownContent = content.substring(endOfFrontMatter + 3).trim();
        } catch (e) {
          console.error('Error parsing front matter:', e);
          this.frontMatter = {};
          this._markdownContent = content;
        }
      } else {
        this.frontMatter = {};
        this._markdownContent = content;
      }
    } else {
      this.frontMatter = {};
      this._markdownContent = content;
    }
  }

  setupPathSegments(): void {
    this.pathSegments = [];

    if (this.filePath) {
      const segments = this.filePath.split('/');
      let currentPath = '';

      for (let i = 0; i < segments.length - 1; i++) {
        if (segments[i]) {
          currentPath += (currentPath ? '/' : '') + segments[i];
          this.pathSegments.push({
            name: segments[i],
            path: currentPath
          });
        }
      }
    }
  }

  onFrontMatterChange(frontMatter: Record<string, any>): void {
    this.frontMatter = frontMatter;
    this.changed = true;
  }

  onFrontMatterInit(frontMatter: Record<string, any>): void {
    this.frontMatter = frontMatter;
  }

  onEditorReady(vditorComponent: any): void {
    this.editor = vditorComponent;
    this.editorRendered = true;
    // Update height after editor is ready
    this.updateEditorHeight();
  }

  calculateEditorHeight(): number {
    // Calculate available height: viewport height minus nav (64px), footer (~53px),
    // file editor header (~80px), front matter section (~100px), and padding/margins (~100px)
    const navHeight = 64;
    const footerHeight = 53;
    const paddingMargins = 100;

    const availableHeight = window.innerHeight - navHeight - footerHeight - paddingMargins;

    // Ensure minimum height of 300px
    return Math.max(300, availableHeight);
  }

  updateEditorHeight(): void {
    const newHeight = this.calculateEditorHeight();
    this.editorOptions = {
      ...this.editorOptions,
      height: newHeight
    };

  }

  saveFile(): void {
    if (this.savingFile) return;

    this.fileError = '';
    this.savingFile = true;
    if (this.editor) {
      this.editor.disabled();
    }

    // Combine front matter and markdown content
    const content = this.generateFileContent();
    let headers = new HttpHeaders();
    if (this.frontMatter) {
      headers = headers.set('X-File-Front-Matter', 'true');
      headers = headers.set('X-File-Front-Matter-Draft', this.frontMatter['draft'] ? 'true' : 'false');
    }

    this.collectionService.updateFileContent(
      this.repositoryId,
      this.collectionName,
      this.filePath,
      content,
      headers
    ).subscribe({
      next: () => {
        this.savingFile = false;
        if (this.editor) {
          this.editor.enable();
        }
        this.changed = false;
        // Navigate back to the collection view
        //this.navigateToCollection();

      },
      error: (err: any) => {
        this.savingFile = false;
        if (this.editor) {
          this.editor.enable();
        }
        const errorMessage = `Failed to save file: ${StrUtils.stringifyHTTPErr(err)}`;
        this.fileError = errorMessage;
        this.showErrorMessage(errorMessage);
      }
    });
  }

  // Helper method to show error messages in a snackbar
  private showErrorMessage(message: string): void {
    this.snackBar.open(message, 'Close', {
      duration: 8000,
      panelClass: ['error-snackbar'],
      verticalPosition: 'top'
    });
  }

  generateFileContent(): string {
    // Only include front matter if it's not empty
    if (Object.keys(this.frontMatter).length === 0) {
      return this.markdownContent;
    }

    const frontMatterYaml = yaml.dump(this.frontMatter);
    return `---\n${frontMatterYaml}---\n\n${this.markdownContent}`;
  }

  cancelEditing(): void {
    this.navigateToCollection();
  }

  navigateToCollection(): void {
    // Navigate back to the collection view

    const pathWithoutFile = this.filePath.substring(0, this.filePath.lastIndexOf('/'));
    this.router.navigate(['/repositories', this.repositoryId, 'collection', this.collectionName], {
      queryParams: {
        path: pathWithoutFile
      }
    });
  }

  canDeactivate(): Observable<boolean> | Promise<boolean> | boolean {
    if (!this.changed) {
      return true;
    }
    const confirmation = window.confirm('You have unsaved changes. Do you really want to leave?');
    return of(confirmation); // Return Observable<boolean>
  }
}
