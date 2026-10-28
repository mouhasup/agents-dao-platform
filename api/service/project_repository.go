package service

import (
	"agents-dao-platform/model"
	"errors"
	"sync"
	"time"

	"gorm.io/gorm"
)

// ProjectRepository définit l'interface de gestion de la persistance des projets
type ProjectRepository interface {
	CreateProject(p *model.Project) (*model.Project, error)
	GetProject(id string) (*model.Project, error)
	ListProjects() ([]*model.Project, error)
	UpdateProject(p *model.Project) (*model.Project, error)
	AddTask(task *model.ProjectTask) error
	ListTasks(projectID string) ([]model.ProjectTask, error)
	UpdateTaskStatus(taskID string, status model.TaskStatus) error
	GetCompanyProfile() (*model.CompanyProfile, error)
	UpdateCompanyProfile(profile *model.CompanyProfile) error
	AddCompanyDocument(doc *model.CompanyDocument) error
	UpdateCompanyDocument(doc *model.CompanyDocument) error
	DeleteCompanyDocument(id string) error
	ListCompanyDocuments() ([]model.CompanyDocument, error)
}

// ==========================================
// IMPLÉMENTATION GORM (PostgreSQL)
// ==========================================

type GormProjectRepository struct {
	db *gorm.DB
}

// NewGormProjectRepository instancie le repository persistant avec GORM & PostgreSQL
func NewGormProjectRepository(db *gorm.DB) ProjectRepository {
	return &GormProjectRepository{db: db}
}

func (r *GormProjectRepository) CreateProject(p *model.Project) (*model.Project, error) {
	if err := r.db.Create(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

func (r *GormProjectRepository) GetProject(id string) (*model.Project, error) {
	var project model.Project
	err := r.db.Preload("Tasks").Preload("Folders").First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *GormProjectRepository) ListProjects() ([]*model.Project, error) {
	var projects []*model.Project
	err := r.db.Preload("Tasks").Preload("Folders").Order("created_at desc").Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *GormProjectRepository) UpdateProject(p *model.Project) (*model.Project, error) {
	err := r.db.Save(p).Error
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *GormProjectRepository) AddTask(task *model.ProjectTask) error {
	return r.db.Create(task).Error
}

func (r *GormProjectRepository) ListTasks(projectID string) ([]model.ProjectTask, error) {
	var tasks []model.ProjectTask
	err := r.db.Where("project_id = ?", projectID).Order("created_at asc").Find(&tasks).Error
	return tasks, err
}

func (r *GormProjectRepository) UpdateTaskStatus(taskID string, status model.TaskStatus) error {
	return r.db.Model(&model.ProjectTask{}).Where("id = ?", taskID).Update("statut", status).Error
}

func (r *GormProjectRepository) GetCompanyProfile() (*model.CompanyProfile, error) {
	var profile model.CompanyProfile
	err := r.db.Preload("Documents").First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			profile = model.CompanyProfile{
				ID:        "default-company",
				Nom:       "Mon Entreprise BTP",
				UpdatedAt: time.Now(),
			}
			if createErr := r.db.Create(&profile).Error; createErr == nil {
				return &profile, nil
			}
		}
		return nil, err
	}
	return &profile, nil
}

func (r *GormProjectRepository) UpdateCompanyProfile(profile *model.CompanyProfile) error {
	return r.db.Save(profile).Error
}

func (r *GormProjectRepository) AddCompanyDocument(doc *model.CompanyDocument) error {
	return r.db.Create(doc).Error
}

func (r *GormProjectRepository) UpdateCompanyDocument(doc *model.CompanyDocument) error {
	return r.db.Save(doc).Error
}

func (r *GormProjectRepository) DeleteCompanyDocument(id string) error {
	return r.db.Delete(&model.CompanyDocument{}, "id = ?", id).Error
}

func (r *GormProjectRepository) ListCompanyDocuments() ([]model.CompanyDocument, error) {
	var docs []model.CompanyDocument
	err := r.db.Order("created_at desc").Find(&docs).Error
	return docs, err
}

// ==========================================
// IMPLÉMENTATION MÉMOIRE (Fallback / Tests)
// ==========================================

type MemoryProjectRepository struct {
	mu             sync.RWMutex
	projects       map[string]*model.Project
	tasks          map[string]*model.ProjectTask
	companyProfile *model.CompanyProfile
	companyDocs    map[string]*model.CompanyDocument
}

func NewMemoryProjectRepository() ProjectRepository {
	return &MemoryProjectRepository{
		projects:    make(map[string]*model.Project),
		tasks:       make(map[string]*model.ProjectTask),
		companyDocs: make(map[string]*model.CompanyDocument),
		companyProfile: &model.CompanyProfile{
			ID:        "cp-default",
			Nom:       "Mon Entreprise BTP & Services",
			NINEA:     "001234567 2V3",
			RCCM:      "SN.DKR.2024.B.12345",
			Email:     "contact@mon-entreprise.sn",
			Telephone: "+221 33 800 00 00",
			Adresse:   "Dakar, Sénégal",
			UpdatedAt: time.Now(),
		},
	}
}

func (r *MemoryProjectRepository) CreateProject(p *model.Project) (*model.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	r.projects[p.ID] = p
	return p, nil
}

func (r *MemoryProjectRepository) GetProject(id string) (*model.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, exists := r.projects[id]
	if !exists {
		return nil, errors.New("projet introuvable")
	}
	var projectTasks []model.ProjectTask
	for _, t := range r.tasks {
		if t.ProjectID == id {
			projectTasks = append(projectTasks, *t)
		}
	}
	p.Tasks = projectTasks
	return p, nil
}

func (r *MemoryProjectRepository) ListProjects() ([]*model.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*model.Project, 0, len(r.projects))
	for _, p := range r.projects {
		list = append(list, p)
	}
	return list, nil
}

func (r *MemoryProjectRepository) UpdateProject(p *model.Project) (*model.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.projects[p.ID]; !exists {
		return nil, errors.New("projet introuvable")
	}
	p.UpdatedAt = time.Now()
	r.projects[p.ID] = p
	return p, nil
}

func (r *MemoryProjectRepository) AddTask(task *model.ProjectTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	task.CreatedAt = time.Now()
	r.tasks[task.ID] = task
	return nil
}

func (r *MemoryProjectRepository) ListTasks(projectID string) ([]model.ProjectTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var tasks []model.ProjectTask
	for _, t := range r.tasks {
		if t.ProjectID == projectID {
			tasks = append(tasks, *t)
		}
	}
	return tasks, nil
}

func (r *MemoryProjectRepository) UpdateTaskStatus(taskID string, status model.TaskStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	task, exists := r.tasks[taskID]
	if !exists {
		return errors.New("tâche introuvable")
	}
	task.Statut = status
	return nil
}

func (r *MemoryProjectRepository) GetCompanyProfile() (*model.CompanyProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.companyProfile, nil
}

func (r *MemoryProjectRepository) UpdateCompanyProfile(profile *model.CompanyProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	profile.UpdatedAt = time.Now()
	r.companyProfile = profile
	return nil
}

func (r *MemoryProjectRepository) AddCompanyDocument(doc *model.CompanyDocument) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()
	r.companyDocs[doc.ID] = doc
	return nil
}

func (r *MemoryProjectRepository) UpdateCompanyDocument(doc *model.CompanyDocument) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	doc.UpdatedAt = time.Now()
	r.companyDocs[doc.ID] = doc
	return nil
}

func (r *MemoryProjectRepository) DeleteCompanyDocument(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.companyDocs, id)
	return nil
}

func (r *MemoryProjectRepository) ListCompanyDocuments() ([]model.CompanyDocument, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var docs []model.CompanyDocument
	for _, d := range r.companyDocs {
		docs = append(docs, *d)
	}
	return docs, nil
}
