import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { RepositoryService, Repository } from '../../services/repository.service';
import { ComponentsModule } from '../../components/components.module';
import {MatInputModule} from '@angular/material/input';
import {StrUtils} from '../../shared/utils/str.utils';
import {NavComponent} from '../../nav/nav.component';

@Component({
  selector: 'app-edit-repository',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterModule, ComponentsModule, MatInputModule, NavComponent],
  templateUrl: './edit-repository.component.html',
  styleUrls: ['./edit-repository.component.scss']
})
export class EditRepositoryComponent implements OnInit {
  repoForm: FormGroup;
  repoId: number;
  repository: Repository | null = null;
  loading = true;
  saving = false;
  error = '';
  branchError = '';
  branches: string[] = [];
  branchesLoading = false;

  constructor(
    private fb: FormBuilder,
    private route: ActivatedRoute,
    private router: Router,
    private repositoryService: RepositoryService
  ) {
    this.repoForm = this.fb.group({
      name: ['', [Validators.required]],
      description: [''],
      branch: ['', [Validators.required]]
    });

    this.repoId = +this.route.snapshot.paramMap.get('id')!;
  }

  ngOnInit(): void {
    this.loadRepository();
  }

  loadRepository(): void {
    this.loading = true;
    this.error = '';

    this.repositoryService.getRepository(this.repoId).subscribe({
      next: (repo) => {
        this.repository = repo;
        this.repoForm.patchValue({
          name: repo.name,
          description: repo.description || '',
          branch: repo.branch || ''
        });
        this.loading = false;
        this.loadBranches();
      },
      error: (err: any) => {
        console.error('Error loading repository:', err);
        this.error = `Failed to load repository. ${StrUtils.stringifyHTTPErr(err)}`;
        this.loading = false;
      }
    });
  }

  loadBranches(): void {
    if (!this.repository) return;

    this.branchesLoading = true;
    this.branchError = '';

    this.repositoryService.getRepositoryBranches(this.repoId).subscribe({
      next: (branches) => {
        this.branches = branches.entries;
        this.branchesLoading = false;
      },
      error: (err: any) => {
        console.error('Error loading branches:', err);
        this.branchError = `Failed to load branches. ${StrUtils.stringifyHTTPErr(err)}`;
        this.branchesLoading = false;
      }
    });
  }

  validateBranch(): void {
    const branch = this.repoForm.get('branch')?.value;
    if (!branch) {
      this.branchError = 'Branch is required';
      return;
    }

    this.branchError = '';
    if (this.branches.length > 0 && !this.branches.includes(branch)) {
      this.branchError = 'Branch does not exist in the repository';
      return;
    }
  }

  onSubmit(): void {
    if (this.repoForm.invalid) {
      return;
    }

    this.validateBranch();
    if (this.branchError) {
      return;
    }

    const newName = this.repoForm.get('name')?.value;
    const newDescription = this.repoForm.get('description')?.value;
    const newBranch = this.repoForm.get('branch')?.value;

    // Check if any values have actually changed
    const nameChanged = newName !== this.repository?.name;
    const descriptionChanged = newDescription !== this.repository?.description;
    const branchChanged = newBranch !== this.repository?.branch;

    // If nothing has changed, just navigate back without saving
    if (!nameChanged && !descriptionChanged && !branchChanged) {
      this.router.navigate(['/home']);
      return;
    }

    this.saving = true;
    // Start with a minimal update object
    const updatedFields: Partial<Repository> = {};

    // Only include fields that have changed
    if (nameChanged) {
      updatedFields.name = newName;
    }

    if (descriptionChanged) {
      updatedFields.description = newDescription;
    }

    if (branchChanged) {
      updatedFields.branch = newBranch;
    }

    const updatedRepo = {
      ...updatedFields
    };

    this.repositoryService.updateRepository(this.repoId, updatedRepo).subscribe({
      next: () => {
        this.saving = false;
        this.router.navigate(['/home']);
      },
      error: (err: any) => {
        console.error('Error updating repository:', err);
        this.error = `Failed to update repository. ${StrUtils.stringifyHTTPErr(err)}`;
        this.saving = false;
      }
    });
  }

  cancel(): void {
    this.router.navigate(['/home']);
  }
}
