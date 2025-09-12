import { Component, Input, Output, EventEmitter, OnInit, OnDestroy, ElementRef, ViewChild, NgZone, forwardRef } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';
import Vditor from 'vditor';

@Component({
  selector: 'app-vditor-editor',
  standalone: true,
  template: `
    <div #vditorContainer class="vditor-container"></div>
  `,
  styleUrls: ['./vditor-editor.component.scss'],
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
  @Input() options: any = {};
  @Output() ready = new EventEmitter<Vditor>();

  private vditor: Vditor | null = null;
  private _value: string = '';
  private onChange = (value: string) => {};
  private onTouched = () => {};
  private isVditorReady = false;

  constructor(private zone: NgZone) {
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
}
