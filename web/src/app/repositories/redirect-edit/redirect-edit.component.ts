import { ChangeDetectionStrategy, Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { CollectionService } from '../../services/collection.service';
import { StrUtils } from '../../shared/utils/str.utils';
import { PageTitleService } from '../../services/page.title.service';

@Component({
  selector: 'app-redirect-edit',
  standalone: true,
  imports: [
    MatCardModule,
    MatIconModule,
    MatProgressSpinnerModule
  ],
  templateUrl: './redirect-edit.component.html',
  changeDetection: ChangeDetectionStrategy.Eager,
  styleUrls: ['./redirect-edit.component.scss']
})
export class RedirectEditComponent implements OnInit {
  isLoading = true;
  error = '';

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private collectionService: CollectionService,
    private pageTitleService: PageTitleService
  ) {}

  ngOnInit(): void {
    this.pageTitleService.title = 'Redirect Edit';
    const repositoryId = this.route.parent?.snapshot.paramMap.get('id') || '';
    const path = this.route.snapshot.queryParamMap.get('path') || '';
    if (!repositoryId || !path) {
      this.error = 'Repository ID and path are required.';
      this.isLoading = false;
      return;
    }

    this.collectionService.resolveConfiguredEditPath(repositoryId, path).subscribe({
      next: (response) => {
        this.router.navigate(
          ['/repositories', repositoryId, 'collection', response.collection, 'edit'],
          {
            queryParams: {path: response.path},
            replaceUrl: true
          }
        );
      },
      error: (err: any) => {
        this.error = `Failed to resolve edit path: ${StrUtils.stringifyHTTPErr(err)}`;
        this.isLoading = false;
      }
    });
  }
}
