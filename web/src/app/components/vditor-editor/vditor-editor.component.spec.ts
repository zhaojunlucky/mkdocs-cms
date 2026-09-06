import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { VditorEditorComponent } from './vditor-editor.component';

describe('VditorEditorComponent find/replace', () => {
  let fixture: ComponentFixture<VditorEditorComponent>;
  let component: VditorEditorComponent;
  let root: HTMLElement;

  function setContent(html: string): void {
    const container = component.vditorContainer.nativeElement as HTMLElement;
    container.innerHTML = `<div class="vditor-wysiwyg"><pre class="vditor-reset" contenteditable="true">${html}</pre></div>`;
    root = container.querySelector('.vditor-reset') as HTMLElement;
  }

  // Matches are painted via the CSS Custom Highlight API (not window.getSelection()),
  // specifically so the highlight survives focus staying in the find input.
  function currentHighlightText(): string {
    const highlight = CSS.highlights.get('vditor-find-match-current');
    const range = highlight ? ([...highlight][0] as Range | undefined) : undefined;
    return range?.toString() ?? '';
  }

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [VditorEditorComponent, NoopAnimationsModule]
    }).compileComponents();

    fixture = TestBed.createComponent(VditorEditorComponent);
    component = fixture.componentInstance;
    // Skip the real (async, CDN-backed) Vditor bootstrap; these tests drive the
    // find/replace logic directly against a hand-built contenteditable pane.
    component.ngOnInit = () => {};
    fixture.detectChanges();

    setContent('Hello <b>world</b>, hello there.');
    (component as unknown as { vditor: unknown }).vditor = {
      getValue: () => root.textContent ?? '',
      setValue: (v: string) => { root.textContent = v; },
      focus: () => {},
      destroy: () => {}
    };
    (component as unknown as { isVditorReady: boolean }).isVditorReady = true;
  });

  it('finds all case-insensitive matches spanning formatted text', () => {
    component.openFind();
    component.findQuery = 'hello';
    component.onFindQueryChange();

    expect(component.matchCount).toBe(2);
    expect(component.currentMatchNumber).toBe(1);
    expect(currentHighlightText().toLowerCase()).toBe('hello');
  });

  it('steps through matches with findNext and wraps around', () => {
    component.openFind();
    component.findQuery = 'hello';
    component.onFindQueryChange();

    component.findNext();
    expect(component.currentMatchNumber).toBe(2);

    component.findNext();
    expect(component.currentMatchNumber).toBe(1);
  });

  it('steps backward with findPrev and wraps around', () => {
    component.openFind();
    component.findQuery = 'hello';
    component.onFindQueryChange();

    component.findPrev();
    expect(component.currentMatchNumber).toBe(2);
  });

  it('respects the match case toggle', () => {
    component.openFind();
    component.findQuery = 'hello';
    component.matchCase = true;
    component.onFindQueryChange();

    expect(component.matchCount).toBe(1);
    expect(currentHighlightText()).toBe('hello');
  });

  it('focuses the find input synchronously as soon as it opens', () => {
    component.openFind();

    expect(component.findInput).toBeTruthy();
    expect(document.activeElement).toBe(component.findInput!.nativeElement);
  });

  it('keeps the find input focused while stepping through matches', () => {
    component.openFind();
    component.findQuery = 'hello';
    component.onFindQueryChange();

    component.findNext();

    expect(document.activeElement).toBe(component.findInput!.nativeElement);
    expect(currentHighlightText().toLowerCase()).toBe('hello');
  });

  it('registers no highlight when the query has no matches', () => {
    component.openFind();
    component.findQuery = 'universe';
    component.onFindQueryChange();
    expect(component.matchCount).toBe(0);
    expect(CSS.highlights.has('vditor-find-match-current')).toBeFalse();
  });

  it('reports no matches for a query that is not present', () => {
    component.openFind();
    component.findQuery = 'xyz-not-there';
    component.onFindQueryChange();

    expect(component.matchCount).toBe(0);
    expect(component.currentMatchNumber).toBe(0);
  });

  it('replaces the currently selected match via execCommand', () => {
    component.openFind();
    component.findQuery = 'world';
    component.replaceQuery = 'universe';
    component.onFindQueryChange();
    expect(component.matchCount).toBe(1);

    component.replaceCurrent();

    expect(root.textContent).toContain('universe');
    expect(root.textContent).not.toContain('world');
  });

  it('replaceAll rewrites the full markdown source via getValue/setValue and notifies the form control', () => {
    const onChange = jasmine.createSpy('onChange');
    component.registerOnChange(onChange);
    (component as unknown as { vditor: { getValue: () => string } }).vditor.getValue = () => 'hello world hello there';

    component.findQuery = 'hello';
    component.replaceQuery = 'hi';
    component.replaceAll();

    expect(onChange).toHaveBeenCalledWith('hi world hi there');
  });

  it('shows a snackbar and makes no change when replaceAll finds nothing', () => {
    const snack = spyOn(component['snack'], 'open');
    const onChange = jasmine.createSpy('onChange');
    component.registerOnChange(onChange);

    component.findQuery = 'not-in-content';
    component.replaceAll();

    expect(onChange).not.toHaveBeenCalled();
    expect(snack).toHaveBeenCalledWith('No matches found', 'Close', { duration: 2000 });
  });

  it('closes and clears match state', () => {
    component.openFind();
    component.findQuery = 'hello';
    component.onFindQueryChange();

    component.closeFindReplace();

    expect(component.findReplaceOpen).toBeFalse();
    expect(component.matchCount).toBe(0);
    expect(component.currentMatchNumber).toBe(0);
    expect(CSS.highlights.has('vditor-find-match')).toBeFalse();
    expect(CSS.highlights.has('vditor-find-match-current')).toBeFalse();
  });
});
