import { Component, OnInit, ViewChild, ElementRef, AfterViewInit, OnDestroy } from '@angular/core';

import {ActivatedRoute, Router, RouterLink, RouterOutlet} from '@angular/router';
import { MatTabsModule } from '@angular/material/tabs';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { RepositoryService } from '../../services/repository.service';
import { Repository, Collection } from '../../services/repository.service';
import {RouteParameterService} from '../../services/routeparameter.service';
import {MatChipsModule} from '@angular/material/chips';
import {StrUtils} from '../../shared/utils/str.utils';
import {PageTitleService} from '../../services/page.title.service';

@Component({
  selector: 'app-repository-detail',
  standalone: true,
  imports: [
    MatTabsModule,
    MatButtonModule,
    MatProgressSpinnerModule,
    RouterLink,
    RouterOutlet,
    MatChipsModule,
    MatIconModule,
    MatTooltipModule
],
  templateUrl: './repository-detail.component.html',
  styleUrls: ['./repository-detail.component.scss']
})
export class RepositoryDetailComponent implements OnInit, AfterViewInit, OnDestroy {
  repository: Repository | null = null;
  collections: Collection[] = [];
  isLoading = true;
  error = '';
  selectedColName: string | null = '';
  showBackToTop = false;
  showScrolltoBottom = false;
  private windowScrollListener: any;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private repositoryService: RepositoryService,
    private routeParameterService: RouteParameterService,
    private pageTitleService: PageTitleService

  ) { }

  ngOnInit(): void {
    this.pageTitleService.title = 'Repository';
    this.route.paramMap.subscribe(params => {
      const id = params.get('id');
      if (id) {
        this.loadRepository(Number(id));
      } else {
        this.error = 'Repository ID is missing';
        this.isLoading = false;
      }
      const collectionName = params.get('collectionName');
      if (collectionName) {
        console.log('collectionName:', collectionName);
      }
    });
    this.routeParameterService.childId$.subscribe(childId => {
      this.selectedColName = childId;
    });
  }

  ngAfterViewInit(): void {
    // Set up scroll listener for the main content div
    setTimeout(() => {
      // Add window scroll listener for mobile devices
      this.windowScrollListener = this.handleWindowScroll.bind(this);
      window.addEventListener('scroll', this.windowScrollListener);
    }, 500);
  }

  ngOnDestroy(): void {
    // Clean up window scroll listener
    if (this.windowScrollListener) {
      window.removeEventListener('scroll', this.windowScrollListener);
    }
  }

  handleScroll(event: Event): void {
    const target = event.target as HTMLElement;
    // Show back-to-top button when scrolled down 300px or more
    const scrollPosition = target.scrollTop;
    const shouldShowBackToTop = scrollPosition > 300;

    // Only update if the value changes to avoid unnecessary renders
    if (this.showBackToTop !== shouldShowBackToTop) {
      this.showBackToTop = shouldShowBackToTop;
    }

    // Check if content is scrollable enough to show scroll-to-bottom button
    const scrollHeight = target.scrollHeight;
    const clientHeight = target.clientHeight;
    const isAtBottom = scrollPosition + clientHeight >= scrollHeight - 20; // 20px threshold
    const shouldShowScrollToBottom = scrollHeight > clientHeight + 300 && !isAtBottom;

    if (this.showScrolltoBottom !== shouldShowScrollToBottom) {
      this.showScrolltoBottom = shouldShowScrollToBottom;
    }

    // For debugging
    // console.log('Content scroll position:', scrollPosition, 'showBackToTop:', this.showBackToTop, 'showScrollToBottom:', this.showScrolltoBottom);
  }

  handleWindowScroll(): void {
    // Show back-to-top button when window is scrolled down 300px or more
    const scrollPosition = window.scrollY || window.pageYOffset;
    const shouldShowBackToTop = scrollPosition > 300;

    // Only update if the value changes to avoid unnecessary renders
    if (this.showBackToTop !== shouldShowBackToTop) {
      this.showBackToTop = shouldShowBackToTop;
    }

    // Check if document is scrollable enough to show scroll-to-bottom button
    const scrollHeight = document.documentElement.scrollHeight;
    const clientHeight = document.documentElement.clientHeight;
    const isAtBottom = scrollPosition + clientHeight >= scrollHeight - 20; // 20px threshold
    const shouldShowScrollToBottom = scrollHeight > clientHeight + 300 && !isAtBottom;

    if (this.showScrolltoBottom !== shouldShowScrollToBottom) {
      this.showScrolltoBottom = shouldShowScrollToBottom;
    }

    // For debugging
    console.log('Window scroll position:', scrollPosition, 'showBackToTop:', this.showBackToTop, 'showScrollToBottom:', this.showScrolltoBottom);
  }

  scrollToTop(): void {
    const mainContentElement = document.querySelector('.main-content');
    if (mainContentElement) {
      mainContentElement.scrollTo({
        top: 0,
        behavior: 'smooth'
      });
    }

    // Also scroll window to top for mobile devices
    window.scrollTo({
      top: 0,
      behavior: 'smooth'
    });
  }

  scrollToBottom() {
    const mainContentElement = document.querySelector('.main-content');
    if (mainContentElement) {
      mainContentElement.scrollTo({
        top: mainContentElement.scrollHeight,
        behavior: 'smooth'
      });
    }

    // Also scroll window to bottom for mobile devices
    window.scrollTo({
      top: document.documentElement.scrollHeight,
      behavior: 'smooth'
    });
  }

  loadRepository(id: number): void {
    this.isLoading = true;
    this.error = '';

    this.repositoryService.getRepository(id).subscribe({
      next: (repo) => {
        this.repository = repo;
        this.pageTitleService.title = `Repository - ${repo.name}`;
        this.loadCollections(id);
      },
      error: (err) => {
        console.error('Error loading repository:', err);
        this.error = `Failed to load repository. ${StrUtils.stringifyHTTPErr(err)}`;
        this.isLoading = false;
      }
    });
  }

  loadCollections(repositoryId: number): void {
    this.repositoryService.getRepositoryCollections(repositoryId).subscribe({
      next: (collections) => {
        this.collections = collections.entries;
        this.isLoading = false;
      },
      error: (err) => {
        console.error('Error loading collections:', err);
        this.error = `Failed to load collections. ${StrUtils.stringifyHTTPErr(err)}`;
        this.isLoading = false;
      }
    });
  }

  selectCollection(collection: Collection) {
    this.router.navigate(['/repositories', this.repository?.id, 'collection', collection.name]);
  }
}
