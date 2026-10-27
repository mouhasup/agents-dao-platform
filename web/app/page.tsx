'use client';

import { useEffect, useState } from "react";
import Link from "next/link";
import { Project, CompanyProfile } from "@/components/dashboard/types";
import NewProjectModal from "@/components/NewProjectModal";
import {
  FolderKanban,
  Plus,
  ExternalLink,
  Clock,
  CheckCircle2,
  AlertCircle,
  FileSearch,
  Building,
  FolderTree,
  FolderCheck,
  Sparkles,
  RefreshCw,
} from "lucide-react";

export default function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [profile, setProfile] = useState<CompanyProfile | null>(null);
  const [initializingWorkspace, setInitializingWorkspace] = useState(false);

  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  const fetchProjects = async () => {
    try {
      setLoading(true);
      const res = await fetch(`${apiUrl}/api/projects`, { credentials: "include" });
      if (res.ok) {
        const data = await res.json();
        setProjects(data || []);
      }
    } catch (err) {
      console.error("Erreur récupération projets:", err);
    } finally {
      setLoading(false);
    }
  };

  const fetchProfile = async () => {
    try {
      const res = await fetch(`${apiUrl}/api/company`, { credentials: "include" });
      if (res.ok) {
        const data = await res.json();
        setProfile(data);
      }
    } catch {
      // Ignorer
    }
  };

  useEffect(() => {
    fetchProjects();
    fetchProfile();
  }, []);

  const handleInitWorkspace = async () => {
    setInitializingWorkspace(true);
    try {
      const res = await fetch(`${apiUrl}/api/workspace/init`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
      });
      if (res.ok) {
        const data = await res.json();
        setProfile(data);
      }
    } catch (err) {
      console.error("Erreur initialisation workspace:", err);
    } finally {
      setInitializingWorkspace(false);
    }
  };

  const handleProjectCreated = (newProject: Project) => {
    setProjects((prev) => [newProject, ...prev]);
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "BROUILLON":
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-slate-100 text-slate-700 border border-slate-200">
            <Clock size={12} className="text-slate-500" />
            Brouillon
          </span>
        );
      case "EN_ANALYSE":
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-200 animate-pulse">
            <AlertCircle size={12} className="text-amber-500" />
            Analyse IA en cours
          </span>
        );
      case "PRET_DEPOT":
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
            <CheckCircle2 size={12} className="text-emerald-600" />
            Prêt pour dépôt
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-blue-50 text-blue-700 border border-blue-200">
            En montage
          </span>
        );
    }
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* En-tête de la page */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-6 border-b border-slate-200">
        <div>
          <h1 className="text-2xl font-black text-slate-900 tracking-tight flex items-center gap-3">
            <FolderKanban className="w-7 h-7 text-blue-600" />
            Dossiers d'Appels d'Offres
          </h1>
          <p className="text-sm text-slate-500 mt-1">
            Gérez vos soumissions, l'arborescence Google Drive et l'analyse multi-agents IA.
          </p>
        </div>

        <button
          onClick={() => setIsModalOpen(true)}
          className="inline-flex items-center justify-center gap-2 px-5 py-2.5 bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold rounded-xl shadow-md shadow-blue-500/20 transition-all cursor-pointer shrink-0"
        >
          <Plus size={18} />
          <span>Nouveau Projet DAO</span>
        </button>
      </div>

      {/* Bannière d'état Workspace Google Drive */}
      {profile && (
        profile.workspace_folder_id ? (
          <div className="bg-white border border-slate-200 rounded-xl px-5 py-3.5 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs shadow-xs">
            <div className="flex items-center gap-2.5">
              <FolderCheck size={18} className="text-emerald-600 shrink-0" />
              <span className="text-slate-600">
                Espace Google Drive connecté : <strong className="text-slate-900 font-mono">AGENTS_DAO_WORKSPACE</strong>
              </span>
              <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
                Prêt
              </span>
            </div>
            {profile.workspace_folder_url && (
              <a
                href={profile.workspace_folder_url}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1.5 font-semibold text-indigo-600 hover:text-indigo-800 shrink-0"
              >
                <span>Ouvrir AGENTS_DAO_WORKSPACE sur Drive</span>
                <ExternalLink size={12} />
              </a>
            )}
          </div>
        ) : (
          <div className="bg-gradient-to-r from-slate-900 via-indigo-950 to-slate-900 text-white rounded-2xl p-5 border border-slate-800 shadow-md flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div className="flex items-start gap-3">
              <div className="p-2.5 bg-indigo-600/30 border border-indigo-500/30 rounded-xl text-indigo-300 shrink-0 mt-0.5">
                <FolderTree size={20} />
              </div>
              <div>
                <p className="text-sm font-bold">Espace Google Drive de l&apos;Entreprise</p>
                <p className="text-xs text-slate-300 mt-0.5 leading-relaxed">
                  L&apos;arborescence principale (<code className="text-indigo-300">AGENTS_DAO_WORKSPACE</code>) regroupant le Coffre-fort RAG et le dossier des projets n&apos;est pas encore créée sur votre Drive.
                </p>
              </div>
            </div>
            <button
              onClick={handleInitWorkspace}
              disabled={initializingWorkspace}
              className="px-4 py-2.5 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white text-xs font-bold rounded-xl transition-all shadow-md shadow-indigo-600/30 flex items-center gap-2 shrink-0 cursor-pointer"
            >
              {initializingWorkspace ? (
                <>
                  <RefreshCw size={14} className="animate-spin" />
                  <span>Création en cours...</span>
                </>
              ) : (
                <>
                  <Sparkles size={14} />
                  <span>Créer le Workspace (1 clic)</span>
                </>
              )}
            </button>
          </div>
        )
      )}

      {/* Contenu principal */}
      {loading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 animate-pulse">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-56 bg-slate-200 rounded-2xl" />
          ))}
        </div>
      ) : projects.length === 0 ? (
        <div className="bg-white border border-slate-200 rounded-2xl p-12 text-center max-w-xl mx-auto shadow-sm">
          <div className="w-16 h-16 bg-blue-50 text-blue-600 rounded-2xl flex items-center justify-center mx-auto mb-4 border border-blue-100">
            <FolderKanban className="w-8 h-8" />
          </div>
          <h3 className="text-lg font-bold text-slate-900">Aucun projet pour le moment</h3>
          <p className="text-sm text-slate-500 mt-1 mb-6">
            Créez votre premier projet d'appel d'offres pour générer automatiquement l'arborescence Drive et lancer les analyses IA.
          </p>
          <button
            onClick={() => setIsModalOpen(true)}
            className="inline-flex items-center gap-2 px-5 py-2.5 bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold rounded-xl shadow-md shadow-blue-500/20 transition-all cursor-pointer"
          >
            <Plus size={18} />
            <span>Créer mon premier projet</span>
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {projects.map((project) => (
            <div
              key={project.id}
              className="bg-white border border-slate-200 hover:border-slate-300 rounded-2xl p-6 shadow-sm hover:shadow-md transition-all flex flex-col justify-between group"
            >
              <div>
                <div className="flex items-center justify-between gap-2 mb-3">
                  {getStatusBadge(project.statut)}
                  <span className="text-xs text-slate-400 font-mono">
                    {new Date(project.created_at).toLocaleDateString("fr-FR")}
                  </span>
                </div>

                <h3 className="text-base font-bold text-slate-900 group-hover:text-blue-600 transition-colors line-clamp-2 leading-snug">
                  {project.titre}
                </h3>

                {project.acheteur && (
                  <p className="text-xs text-slate-500 flex items-center gap-1.5 mt-2">
                    <Building size={14} className="text-slate-400 shrink-0" />
                    <span className="truncate">{project.acheteur}</span>
                  </p>
                )}

                {project.folders && project.folders.length > 0 && (
                  <div className="mt-4 pt-3 border-t border-slate-100 flex items-center gap-1.5 text-xs text-slate-500">
                    <FolderTree size={14} className="text-indigo-500" />
                    <span>Arborescence Drive créée ({project.folders.length} dossiers)</span>
                  </div>
                )}
              </div>

              <div className="mt-6 pt-4 border-t border-slate-100 flex items-center justify-between gap-3">
                {project.drive_folder_url ? (
                  <a
                    href={project.drive_folder_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1 text-xs font-semibold text-slate-600 hover:text-blue-600 transition-colors"
                  >
                    <span>Ouvrir Drive</span>
                    <ExternalLink size={12} />
                  </a>
                ) : (
                  <span className="text-xs text-slate-400">Drive local</span>
                )}

                <Link
                  href={`/projects/${project.id}`}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-slate-900 hover:bg-slate-800 text-white text-xs font-bold rounded-lg shadow-sm transition-colors"
                >
                  <FileSearch size={14} />
                  <span>Détail DAO</span>
                </Link>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Modal de création de projet */}
      <NewProjectModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onCreated={handleProjectCreated}
      />
    </div>
  );
}
