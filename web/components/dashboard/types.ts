export interface Lot {
  numero_lot: number;
  intitule: string;
  cautionnement: string;
  criteres_techniques: string[];
  pieces_requises: string[];
}

export interface DataAnalyseDAO {
  acheteur: string;
  objet_marche: string;
  numero_appel_offres?: string;
  criteres_eliminatoires?: string[];
  recommandations?: string;
  lots: Lot[];
}

export type ProjectStatus =
  | "BROUILLON"
  | "EN_ANALYSE"
  | "EN_MONTAGE"
  | "PRET_DEPOT"
  | "DEPOSE"
  | "GAGNE"
  | "PERDU";

export interface ProjectTask {
  id: string;
  project_id: string;
  titre: string;
  description: string;
  categorie: "ADMINISTRATIF" | "TECHNIQUE" | "FINANCIER" | "VISITE_SITE";
  echeance?: string;
  priorite: "HAUTE" | "MOYENNE" | "BASSE";
  statut: "A_FAIRE" | "EN_COURS" | "TERMINEE";
}

export interface DriveFolder {
  id: string;
  project_id: string;
  google_folder_id: string;
  nom: string;
  categorie: string;
}

export interface Project {
  id: string;
  titre: string;
  numero_dao?: string;
  acheteur?: string;
  statut: ProjectStatus;
  drive_folder_id?: string;
  drive_folder_url?: string;
  date_limite_depot?: string;
  analyse_result_json?: string;
  analyse_error?: string;
  analysed_at?: string;
  created_at: string;
  updated_at: string;
  tasks?: ProjectTask[];
  folders?: DriveFolder[];
}

export interface AdminAnalysis {
  acheteur: string;
  objet_marche: string;
  numero_appel_offres?: string;
  date_limite_depot?: string;
  criteres_eliminatoires?: string[];
  pieces_administratives?: string[];
  calendrier?: { etape: string; date?: string }[];
  lots?: { numero_lot: number; intitule: string; cautionnement?: string; pieces_requises?: string[] }[];
  recommandations?: string;
}

export interface TechnicalAnalysis {
  consistance_travaux: string;
  contraintes_execution?: string[];
  normes_exigees?: string[];
  materiel_requis?: string[];
  personnel_cle?: string[];
  plan_memoire_technique?: string[];
  points_vigilance?: string;
}

export interface ConsolidatedAnalysis {
  documents: string[];
  administrative?: AdminAnalysis;
  technique?: TechnicalAnalysis;
  erreurs?: string[];
}

export interface DriveDocument {
  id: string;
  name: string;
  mimeType: string;
  size?: string;
  webViewLink?: string;
  createdTime?: string;
}

export type TypeDocumentEntreprise =
  | "CV_EXPERT"
  | "ATTESTATION_FISCALE_SOCIALE"
  | "REFERENCE_SERVICE_FAIT"
  | "JUSTIFICATIF_MATERIEL"
  | "BILAN_FINANCIER"
  | "AUTRE";

export interface CompanyDocument {
  id: string;
  type_doc: TypeDocumentEntreprise;
  titre: string;
  description: string;
  google_drive_id?: string;
  web_view_link?: string;
  file_name?: string;
  date_expiration?: string;
  metadonnees_json?: string;
  created_at: string;
}

export interface CompanyProfile {
  id: string;
  nom: string;
  ninea: string;
  rccm: string;
  email: string;
  telephone: string;
  adresse: string;
  workspace_folder_id?: string;
  workspace_folder_url?: string;
  rag_folder_id?: string;
  projects_folder_id?: string;
  documents?: CompanyDocument[];
  updated_at: string;
}

export interface RequirementMatch {
  requirement: string;
  category: string;
  is_eliminatory: boolean;
  status: "CONFORME" | "A_RENOUVELER" | "MANQUANT";
  matched_document?: CompanyDocument;
  comment: string;
  matching_task_id?: string;
}

export interface ComplianceCheckResult {
  project_id: string;
  global_status: "CONFORME" | "ATTENTION" | "NON_CONFORME";
  total_requirements: number;
  matched_count: number;
  missing_count: number;
  expired_count: number;
  requirements: RequirementMatch[];
  company_documents_count: number;
  offre_admin_folder_id?: string;
  offre_admin_folder_url?: string;
}

export interface MemoireSection {
  id: string;
  titre: string;
  contenu: string;
  statut: "A_REDIGER" | "GENERE" | "VALIDE";
  updated_at?: string;
}

export interface PersonnelAffectation {
  role_requis: string;
  nom: string;
  experience_annees: number;
  qualifications: string;
  document_id?: string;
  document_url?: string;
}

export interface MaterielMobilisation {
  type_requis: string;
  designation: string;
  immatriculation?: string;
  statut: "PROPRIETE" | "LOCATION" | "LEASING";
  document_id?: string;
  document_url?: string;
}

export interface TechnicalOfferData {
  project_id: string;
  titre: string;
  sections: MemoireSection[];
  personnel: PersonnelAffectation[];
  materiel: MaterielMobilisation[];
  drive_file_id?: string;
  drive_folder_url?: string;
  last_exported_at?: string;
}
