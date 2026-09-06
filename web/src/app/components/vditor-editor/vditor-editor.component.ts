import { ChangeDetectorRef, Component, Input, Output, EventEmitter, OnInit, OnDestroy, ElementRef, ViewChild, NgZone, forwardRef, ChangeDetectionStrategy } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR, FormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatTooltipModule } from '@angular/material/tooltip';
import Vditor from 'vditor';

@Component({
  selector: 'app-vditor-editor',
  standalone: true,
  imports: [FormsModule, MatButtonModule, MatIconModule, MatSnackBarModule, MatTooltipModule],
  templateUrl: './vditor-editor.component.html',
  styleUrls: ['./vditor-editor.component.scss'],
  changeDetection: ChangeDetectionStrategy.Eager,
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => VditorEditorComponent),
      multi: true
    }
  ]
})
export class VditorEditorComponent implements OnInit, OnDestroy, ControlValueAccessor {
  @ViewChild('vditorContainer', { static: true }) vditorContainer!: ElementRef;
  @ViewChild('findInput') findInput?: ElementRef<HTMLInputElement>;
  @Input() options: any = {};
  @Output() ready = new EventEmitter<Vditor>();

  findReplaceOpen = false;
  replaceRowOpen = false;
  findQuery = '';
  replaceQuery = '';
  matchCase = false;
  matchCount = 0;
  currentMatchNumber = 0;

  private vditor: Vditor | null = null;
  private _value: string = '';
  private onChange = (value: string) => {};
  private onTouched = () => {};
  private isVditorReady = false;
  private matches: Range[] = [];
  private currentMatchIndex = -1;
  private findReplaceKeyboardListener: ((e: KeyboardEvent) => void) | null = null;
  private static readonly MATCH_HIGHLIGHT = 'vditor-find-match';
  private static readonly CURRENT_MATCH_HIGHLIGHT = 'vditor-find-match-current';
  private readonly highlightsSupported = typeof CSS !== 'undefined' && 'highlights' in CSS;

  constructor(private zone: NgZone, private cdr: ChangeDetectorRef, private snack: MatSnackBar) {
    // Add global error handler to catch DOM errors from Vditor
    this.addGlobalErrorHandler();
  }

  private addGlobalErrorHandler(): void {
    // Add window error handler to catch runtime DOM errors
    window.addEventListener('error', (event: ErrorEvent): boolean => {
      if (event.message && event.message.includes('classList')) {
        console.warn('Caught Vditor classList error, suppressing:', event.message);
        event.preventDefault();
        return false;
      }
      return true;
    });

    // Also handle unhandled promise rejections
    window.addEventListener('unhandledrejection', (event: PromiseRejectionEvent): void => {
      if (event.reason && event.reason.message && event.reason.message.includes('classList')) {
        console.warn('Caught Vditor classList promise rejection, suppressing:', event.reason.message);
        event.preventDefault();
      }
    });
  }

  private patchDOMAccess(): void {
    // Monkey patch Element prototype to handle null classList access
    const originalClassList = Object.getOwnPropertyDescriptor(Element.prototype, 'classList');
    if (originalClassList) {
      Object.defineProperty(Element.prototype, 'classList', {
        get: function() {
          try {
            return originalClassList.get?.call(this) || {
              add: () => {},
              remove: () => {},
              contains: () => false,
              toggle: () => false,
              replace: () => false
            };
          } catch (e) {
            console.warn('classList access error caught and handled');
            return {
              add: () => {},
              remove: () => {},
              contains: () => false,
              toggle: () => false,
              replace: () => false
            };
          }
        },
        configurable: true
      });
    }
  }

  ngOnInit(): void {
    // Run Vditor initialization outside Angular's zone to avoid change detection issues
    this.zone.runOutsideAngular(() => {
      setTimeout(() => {
        this.initVditor();
      }, 200); // Increased delay to ensure DOM is fully ready
    });
  }

  ngOnDestroy(): void {
    this.isVditorReady = false;
    this.detachFindReplaceKeyboardListener();
    this.matches = [];
    this.currentMatchIndex = -1;
    this.syncHighlights();
    if (this.vditor) {
      try {
        this.vditor.destroy();
      } catch (error) {
        console.warn('Error destroying Vditor:', error);
      }
      this.vditor = null;
    }
  }

  private initVditor(): void {
    if (!this.vditorContainer?.nativeElement) {
      console.warn('Vditor container not available');
      return;
    }

    // Monkey patch to prevent classList errors
    this.patchDOMAccess();

    const defaultOptions = {
      theme: 'classic',
      language: 'markdown',
      lang: 'en_US',
      icon: 'material',
      mode: 'sv',
      cdn: '/assets/vditor',
      tab: '    ',
      counter: {
        enable: true,
      },
      cache: {
        enable: false,
      },
      customWysiwygToolbar: () => {
        // Required function for Vditor 3.11.2+
        return [];
      },
      preview: {
        mode: 'editor',  // Only show editor, disable preview to avoid DOM issues
        hljs: {
          lineNumber: true,
          enable: true,
          style: 'github'
        },
        markdown: {
          toc: false,  // Disable table of contents
          sanitize: true,
          codeBlockPreview: false,
          mathBlockPreview: false,
          paragraphBeginningSpace: false,
          autoSpace: true,
          listStyle: true,
        },
        actions: []  // Remove all preview actions
      },
      height: 400,
      value: this._value,
      input: (value: string) => {
        this.zone.run(() => {
          this._value = value;
          this.onChange(value);
          this.onTouched();
        });
      },
      after: () => {
        this.zone.run(() => {
          this.isVditorReady = true;
          // Set initial value if it exists
          if (this._value && this.vditor) {
            this.vditor.setValue(this._value);
          }
          this.ready.emit(this.vditor!);
          if (this.options.after) {
            this.options.after();
          }
          this.attachFindReplaceKeyboardListener();
        });
      }
    };

    const mergedOptions = { ...defaultOptions, ...this.options };

    try {
      this.vditor = new Vditor(this.vditorContainer.nativeElement, mergedOptions);
    } catch (error) {
      console.error('Error initializing Vditor:', error);
      this.zone.run(() => {
        // Emit ready event even if there's an error to prevent loading overlay from hanging
        this.ready.emit(null as any);
      });
    }
  }

  // ControlValueAccessor implementation
  writeValue(value: string): void {
    this._value = value || '';
    if (this.vditor && this.isVditorReady) {
      this.vditor.setValue(this._value);
      if (this.findReplaceOpen) {
        // external content update: match ranges may point at replaced DOM nodes
        this.runSearch(this.currentMatchIndex);
      }
    }
  }

  registerOnChange(fn: (value: string) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.onTouched = fn;
  }

  setDisabledState(isDisabled: boolean): void {
    if (this.vditor && this.isVditorReady) {
      if (isDisabled) {
        this.vditor.disabled();
      } else {
        this.vditor.enable();
      }
    }
  }

  // Public methods to control the editor
  getValue(): string {
    return this.vditor ? this.vditor.getValue() : this._value;
  }

  setValue(value: string): void {
    this._value = value;
    if (this.vditor && this.isVditorReady) {
      this.vditor.setValue(value);
    }
  }

  disabled(): void {
    if (this.vditor && this.isVditorReady) {
      this.vditor.disabled();
    }
  }

  enable(): void {
    if (this.vditor && this.isVditorReady) {
      this.vditor.enable();
    }
  }

  getVditor(): Vditor | null {
    return this.vditor;
  }

  // Find/replace
  openFind(): void {
    if (!this.vditor || !this.isVditorReady) return;
    this.findReplaceOpen = true;
    this.replaceRowOpen = false;
    this.cdr.detectChanges();
    this.findInput?.nativeElement.focus();
    this.runSearch();
  }

  closeFindReplace(): void {
    this.findReplaceOpen = false;
    this.replaceRowOpen = false;
    this.matches = [];
    this.currentMatchIndex = -1;
    this.matchCount = 0;
    this.currentMatchNumber = 0;
    this.syncHighlights();
    this.vditor?.focus();
    this.cdr.markForCheck();
  }

  toggleMatchCase(): void {
    this.matchCase = !this.matchCase;
    this.runSearch();
  }

  onFindQueryChange(): void {
    this.runSearch();
  }

  onFindEnter(event: Event): void {
    const shiftKey = event instanceof KeyboardEvent && event.shiftKey;
    shiftKey ? this.findPrev() : this.findNext();
  }

  findNext(): void {
    if (this.matches.length === 0) return;
    this.currentMatchIndex = (this.currentMatchIndex + 1) % this.matches.length;
    this.selectCurrentMatch();
  }

  findPrev(): void {
    if (this.matches.length === 0) return;
    this.currentMatchIndex = (this.currentMatchIndex - 1 + this.matches.length) % this.matches.length;
    this.selectCurrentMatch();
  }

  replaceCurrent(): void {
    if (!this.vditor || this.currentMatchIndex < 0 || this.currentMatchIndex >= this.matches.length) return;
    const selection = window.getSelection();
    if (!selection) return;
    selection.removeAllRanges();
    selection.addRange(this.matches[this.currentMatchIndex]);
    // Vditor's wysiwyg pane listens for the native input event execCommand fires,
    // and syncs the markdown source through the same `input` option as normal typing.
    document.execCommand('insertText', false, this.replaceQuery);
    queueMicrotask(() => this.runSearch(this.currentMatchIndex));
  }

  replaceAll(): void {
    if (!this.vditor || !this.isVditorReady || !this.findQuery) return;
    const pattern = this.buildRegExp();
    if (!pattern) return;
    const source = this.vditor.getValue();
    const count = (source.match(pattern) ?? []).length;
    if (count === 0) {
      this.snack.open('No matches found', 'Close', { duration: 2000 });
      return;
    }
    const next = source.replace(pattern, this.replaceQuery);
    // setValue does not fire the input callback, so notify the form control directly
    this._value = next;
    this.vditor.setValue(next);
    this.onChange(next);
    this.onTouched();
    this.snack.open(`Replaced ${count} occurrence${count === 1 ? '' : 's'}`, 'Close', { duration: 2500 });
    this.runSearch();
  }

  private attachFindReplaceKeyboardListener(): void {
    const container = this.vditorContainer?.nativeElement;
    if (!container) return;
    this.findReplaceKeyboardListener = (e: KeyboardEvent) => {
      const isFindShortcut = (e.ctrlKey || e.metaKey) && !e.shiftKey && e.key.toLowerCase() === 'f';
      if (isFindShortcut) {
        e.preventDefault();
        this.zone.run(() => this.openFind());
        return;
      }
      if (e.key === 'Escape' && this.findReplaceOpen) {
        e.preventDefault();
        this.zone.run(() => this.closeFindReplace());
      }
    };
    container.addEventListener('keydown', this.findReplaceKeyboardListener, true);
  }

  private detachFindReplaceKeyboardListener(): void {
    const container = this.vditorContainer?.nativeElement;
    if (container && this.findReplaceKeyboardListener) {
      container.removeEventListener('keydown', this.findReplaceKeyboardListener, true);
    }
    this.findReplaceKeyboardListener = null;
  }

  private getEditableRoot(): HTMLElement | null {
    const container = this.vditorContainer?.nativeElement as HTMLElement | undefined;
    if (!container) return null;
    return container.querySelector<HTMLElement>('.vditor-wysiwyg .vditor-reset')
      ?? container.querySelector<HTMLElement>('.vditor-ir .vditor-reset')
      ?? null;
  }

  private buildRegExp(): RegExp | null {
    if (!this.findQuery) return null;
    const escaped = this.findQuery.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    return new RegExp(escaped, this.matchCase ? 'g' : 'gi');
  }

  private runSearch(preferredIndex = -1): void {
    this.matches = [];
    this.currentMatchIndex = -1;
    const root = this.getEditableRoot();
    const pattern = this.buildRegExp();
    if (!root || !pattern) {
      this.matchCount = 0;
      this.currentMatchNumber = 0;
      this.syncHighlights();
      this.cdr.markForCheck();
      return;
    }

    const chunks: { node: Text; start: number }[] = [];
    let buffer = '';
    const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
    let node: Node | null;
    while ((node = walker.nextNode())) {
      const text = node.textContent ?? '';
      if (!text) continue;
      chunks.push({ node: node as Text, start: buffer.length });
      buffer += text;
    }

    let match: RegExpExecArray | null;
    while ((match = pattern.exec(buffer))) {
      if (match[0].length === 0) { pattern.lastIndex++; continue; }
      const range = this.buildRangeFromOffsets(chunks, match.index, match.index + match[0].length);
      if (range) this.matches.push(range);
    }

    this.matchCount = this.matches.length;
    if (this.matches.length > 0) {
      this.currentMatchIndex = preferredIndex >= 0 ? Math.min(preferredIndex, this.matches.length - 1) : 0;
      this.selectCurrentMatch();
    } else {
      this.currentMatchNumber = 0;
      this.syncHighlights();
      this.cdr.markForCheck();
    }
  }

  private buildRangeFromOffsets(chunks: { node: Text; start: number }[], start: number, end: number): Range | null {
    const startChunk = this.findChunkForOffset(chunks, start, false);
    const endChunk = this.findChunkForOffset(chunks, end, true);
    if (!startChunk || !endChunk) return null;
    const range = document.createRange();
    range.setStart(startChunk.node, start - startChunk.start);
    range.setEnd(endChunk.node, end - endChunk.start);
    return range;
  }

  private findChunkForOffset(chunks: { node: Text; start: number }[], offset: number, isEnd: boolean): { node: Text; start: number } | null {
    for (const chunk of chunks) {
      const chunkEnd = chunk.start + (chunk.node.textContent?.length ?? 0);
      if (isEnd ? offset <= chunkEnd : offset < chunkEnd) return chunk;
    }
    return chunks.length > 0 ? chunks[chunks.length - 1] : null;
  }

  private selectCurrentMatch(): void {
    if (this.currentMatchIndex < 0 || this.currentMatchIndex >= this.matches.length) return;
    this.syncHighlights();
    const range = this.matches[this.currentMatchIndex];
    const container = range.startContainer.nodeType === Node.TEXT_NODE
      ? range.startContainer.parentElement
      : range.startContainer as HTMLElement;
    container?.scrollIntoView({ block: 'center', behavior: 'smooth' });
    this.currentMatchNumber = this.currentMatchIndex + 1;
    this.cdr.markForCheck();
  }

  // Highlights matches via the CSS Custom Highlight API rather than window.getSelection():
  // a real Selection lives in only one focused editing host at a time, so painting matches
  // that way would collapse the moment focus stayed in (or returned to) the find input.
  private syncHighlights(): void {
    if (!this.highlightsSupported) return;
    const registry = CSS.highlights;
    if (this.matches.length === 0) {
      registry.delete(VditorEditorComponent.MATCH_HIGHLIGHT);
      registry.delete(VditorEditorComponent.CURRENT_MATCH_HIGHLIGHT);
      return;
    }
    const others = this.matches.filter((_, i) => i !== this.currentMatchIndex);
    if (others.length > 0) {
      registry.set(VditorEditorComponent.MATCH_HIGHLIGHT, new Highlight(...others));
    } else {
      registry.delete(VditorEditorComponent.MATCH_HIGHLIGHT);
    }
    if (this.currentMatchIndex >= 0) {
      registry.set(VditorEditorComponent.CURRENT_MATCH_HIGHLIGHT, new Highlight(this.matches[this.currentMatchIndex]));
    } else {
      registry.delete(VditorEditorComponent.CURRENT_MATCH_HIGHLIGHT);
    }
  }
}
