'use client';

import { useState } from "react";
import {
  TechnicalOfferData,
  MemoireSection,
  PersonnelAffectation,
  MaterielMobilisation,
  CompanyDocument,
} from "./types";
import {
  FileText,
  Sparkles,
  RefreshCw,
  Check,
  CheckCircle2,
  FolderTree,
  FolderCheck,
  ExternalLink,
  Users,
  Truck,
  Edit3,
  Eye,
  Save,
  Send,
  AlertCircle,
  FileCheck2,
  ChevronRight,
  Shield,
  Layers,
  Download,
  Building2,
  UserCheck,
} from "lucide-react";

interface TechnicalOfferEditorProps {
  projectId: string;
  data: TechnicalOfferData;
  companyDocs: CompanyDocument[];
  onRefresh: () => Promise<void>;
}

export default function TechnicalOfferEditor({
  projectId,
  data,
  companyDocs,
  onRefresh,
}: TechnicalOfferEditorProps) {
  const sections = data?.sections || [];
  const personnel = data?.personnel || [];
  const materiel = data?.materiel || [];

  const [activeSubTab, setActiveSubTab] = useState<"chapters" | "personnel" | "materiel">("chapters");
  const [selectedSectionId, setSelectedSectionId] = useState<string>(
    sections.length > 0 ? sections[0].id : ""
  );

  // Édition de section
  const [customPrompt, setCustomPrompt] = useState("");
  const [generatingSection, setGeneratingSection] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [currentContent, setCurrentContent] = useState("");
  const [savingOffer, setSavingOffer] = useState(false);
  const [saveMessage, setSaveMessage] = useState<string | null>(null);

  // Export Drive
  const [exportingDrive, setExportingDrive] = useState(false);
  const [exportResult, setExportResult] = useState<{ text: string; url?: string; error?: boolean } | null>(null);

  // Proposition & Validation IA Personnel / Matériel
  const [proposingPersonnel, setProposingPersonnel] = useState(false);
  const [proposingMateriel, setProposingMateriel] = useState(false);
  const [validatedPersonnel, setValidatedPersonnel] = useState<number[]>([]);
  const [validatedMateriel, setValidatedMateriel] = useState<number[]>([]);
  const [copyingCVs, setCopyingCVs] = useState(false);
  const [cvCopySuccessMsg, setCvCopySuccessMsg] = useState<string | null>(null);

  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  const selectedSection = sections.find((s) => s.id === selectedSectionId) || sections[0];

  const handleSelectSection = (sec: MemoireSection) => {
    setSelectedSectionId(sec.id);
    setCurrentContent(sec.contenu);
    setEditMode(false);
    setSaveMessage(null);
  };

  const handleGenerateSection = async () => {
    if (!selectedSection) return;
    setGeneratingSection(true);
    setSaveMessage(null);
    try {
      const res = await fetch(`${apiUrl}/api/projects/${projectId}/technical-offer/generate-section`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          section_id: selectedSection.id,
          custom_instruction: customPrompt.trim(),
        }),
      });
      const updatedSection = await res.json();
      if (res.ok) {
        setCurrentContent(updatedSection.contenu);
        await onRefresh();
        setSaveMessage("Section rédigée avec succès par l'IA !");
      } else {
        setSaveMessage(updatedSection.error || "Erreur lors de la rédaction IA.");
      }
    } catch {
      setSaveMessage("Impossible de joindre l'API.");
    } finally {
      setGeneratingSection(false);
    }
  };

  const handleSaveSections = async (updatedSections: MemoireSection[]) => {
    setSavingOffer(true);
    setSaveMessage(null);
    try {
      const payload: TechnicalOfferData = {
        ...data,
        sections: updatedSections,
      };
      const res = await fetch(`${apiUrl}/api/projects/${projectId}/technical-offer/save`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify(payload),
      });
      if (res.ok) {
        await onRefresh();
        setSaveMessage("Modifications enregistrées !");
        setEditMode(false);
      }
    } catch {
      setSaveMessage("Erreur d'enregistrement.");
    } finally {
      setSavingOffer(false);
    }
  };

  const handleSaveCurrentSectionText = () => {
    const updated = sections.map((s) =>
      s.id === selectedSection?.id ? { ...s, contenu: currentContent } : s
    );
    handleSaveSections(updated);
  };

  const handleValidateSection = () => {
    const updated = sections.map((s) =>
      s.id === selectedSection?.id ? { ...s, statut: "VALIDE" as const } : s
    );
    handleSaveSections(updated);
  };

  const handleUpdatePersonnel = async (index: number, updatedItem: PersonnelAffectation) => {
    const newPersonnel = [...personnel];
    newPersonnel[index] = updatedItem;
    const payload = { ...data, personnel: newPersonnel };
    await fetch(`${apiUrl}/api/projects/${projectId}/technical-offer/save`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify(payload),
    });
    await onRefresh();
  };

  const handleUpdateMateriel = async (index: number, updatedItem: MaterielMobilisation) => {
    const newMateriel = [...materiel];
    newMateriel[index] = updatedItem;
    const payload = { ...data, materiel: newMateriel };
    await fetch(`${apiUrl}/api/projects/${projectId}/technical-offer/save`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify(payload),
    });
    await onRefresh();
  };

  const handleExportToDrive = async () => {
    setExportingDrive(true);
    setExportResult(null);
    try {
      const res = await fetch(`${apiUrl}/api/projects/${projectId}/technical-offer/export-drive`, {
        method: "POST",
        credentials: "include",
      });
      const resData = await res.json();
      if (res.ok) {
        setExportResult({
          text: resData.message || "Mémoire technique exporté avec succès sur Drive !",
          url: resData.file_url,
        });
        await onRefresh();
      } else {
        setExportResult({
          text: resData.error || "Erreur lors de l'export sur Google Drive.",
          error: true,
        });
      }
    } catch {
      setExportResult({
        text: "Impossible de joindre l'API pour l'export Drive.",
        error: true,
      });
    } finally {
      setExportingDrive(false);
    }
  };

  const handleProposePersonnel = async () => {
    setProposingPersonnel(true);
    setSaveMessage(null);
    try {
      const res = await fetch(`${apiUrl}/api/projects/${projectId}/technical-offer/propose-personnel`, {
        method: "POST",
        credentials: "include",
      });
      const resData = await res.json();
      if (res.ok && resData.personnel) {
        await onRefresh();
        setSaveMessage("L'agent IA a proposé et pré-rempli l'ensemble des profils d'experts ! Vous pouvez ajuster chaque champ et valider les profils.");
      } else {
        setSaveMessage(resData.error || "Erreur lors de la proposition du personnel.");
      }
    } catch {
      setSaveMessage("Impossible de joindre l'API.");
    } finally {
      setProposingPersonnel(false);
    }
  };

  const handleProposeMateriel = async () => {
    setProposingMateriel(true);
    setSaveMessage(null);
    try {
      const res = await fetch(`${apiUrl}/api/projects/${projectId}/technical-offer/propose-materiel`, {
        method: "POST",
        credentials: "include",
      });
      const resData = await res.json();
      if (res.ok && resData.materiel) {
        await onRefresh();
        setSaveMessage("L'agent IA a proposé et complété les équipements et engins de chantier adaptés au projet !");
      } else {
        setSaveMessage(resData.error || "Erreur lors de la proposition du matériel.");
      }
    } catch {
      setSaveMessage("Impossible de joindre l'API.");
    } finally {
      setProposingMateriel(false);
    }
  };

  const toggleValidatePersonnel = (idx: number) => {
    setValidatedPersonnel((prev) =>
      prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]
    );
  };

  const validateAllPersonnel = () => {
    setValidatedPersonnel(personnel.map((_, i) => i));
  };

  const toggleValidateMateriel = (idx: number) => {
    setValidatedMateriel((prev) =>
      prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]
    );
  };

  const validateAllMateriel = () => {
    setValidatedMateriel(materiel.map((_, i) => i));
  };

  const handleCopyCVs = async () => {
    setCopyingCVs(true);
    setCvCopySuccessMsg(null);
    try {
      const res = await fetch(`${apiUrl}/api/projects/${projectId}/technical-offer/copy-cvs`, {
        method: "POST",
        credentials: "include",
      });
      const resData = await res.json();
      if (res.ok) {
        setCvCopySuccessMsg(resData.message);
        await onRefresh();
      } else {
        alert(resData.error || "Erreur lors du transfert des CVs.");
      }
    } catch {
      alert("Impossible de joindre l'API.");
    } finally {
      setCopyingCVs(false);
    }
  };

  const readySectionsCount = sections.filter((s) => s.statut === "GENERE" || s.statut === "VALIDE").length;

  return (
    <div className="space-y-6">
      {/* Barre d'en-tête de l'Offre Technique */}
      <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm space-y-4">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2">
              <span className="text-xs font-bold font-mono text-indigo-600 bg-indigo-50 px-2 py-0.5 rounded uppercase">
                Offre Technique & CCTP
              </span>
              <span className="text-[11px] font-semibold text-slate-500">
                {readySectionsCount}/{sections.length} chapitres rédigés
              </span>
            </div>
            <h2 className="text-lg font-black text-slate-900 mt-1">{data.titre}</h2>
            <p className="text-xs text-slate-500 mt-0.5">
              Rédigez la méthodologie d'exécution, affectez le personnel clé et exportez le mémoire technique au format <b>Microsoft Word (.docx)</b>.
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-2 sm:gap-3 shrink-0">
            {data.drive_folder_url && (
              <a
                href={data.drive_folder_url}
                target="_blank"
                rel="noopener noreferrer"
                className="px-3.5 py-2.5 bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-semibold rounded-xl flex items-center gap-1.5 transition-colors"
              >
                <FolderTree size={14} className="text-indigo-600" />
                <span>Voir sur Drive</span>
                <ExternalLink size={12} />
              </a>
            )}

            <a
              href={`${apiUrl}/api/projects/${projectId}/technical-offer/download`}
              target="_blank"
              rel="noopener noreferrer"
              download
              title="Télécharger le Mémoire Technique complet au format Word .docx"
              className="px-3.5 py-2.5 bg-blue-50 hover:bg-blue-100 text-blue-700 border border-blue-200 text-xs font-bold rounded-xl flex items-center gap-1.5 shadow-sm transition-colors cursor-pointer"
            >
              <Download size={14} className="text-blue-600" />
              <span>Télécharger Word (.docx)</span>
            </a>

            <button
              onClick={handleExportToDrive}
              disabled={exportingDrive}
              className="px-5 py-2.5 bg-gradient-to-r from-indigo-600 to-blue-600 hover:from-indigo-700 hover:to-blue-700 text-white text-xs font-bold rounded-xl shadow-md shadow-indigo-500/20 disabled:opacity-50 transition-all flex items-center gap-2 cursor-pointer"
            >
              {exportingDrive ? (
                <>
                  <RefreshCw size={14} className="animate-spin" />
                  <span>Exportation Word sur Drive...</span>
                </>
              ) : (
                <>
                  <FileCheck2 size={15} />
                  <span>Exporter Word vers 03_Offre_Technique</span>
                </>
              )}
            </button>
          </div>
        </div>

        {/* Message d'export */}
        {exportResult && (
          <div
            className={`p-3.5 rounded-xl text-xs flex items-center justify-between gap-3 ${
              exportResult.error
                ? "bg-rose-50 text-rose-800 border border-rose-200"
                : "bg-emerald-50 text-emerald-800 border border-emerald-200"
            }`}
          >
            <div className="flex items-center gap-2">
              {exportResult.error ? <AlertCircle size={16} /> : <CheckCircle2 size={16} />}
              <span>{exportResult.text}</span>
            </div>
            {exportResult.url && (
              <a
                href={exportResult.url}
                target="_blank"
                rel="noopener noreferrer"
                className="underline font-bold hover:text-emerald-950 flex items-center gap-1 shrink-0"
              >
                <span>Consulter le fichier sur Google Drive</span>
                <ExternalLink size={12} />
              </a>
            )}
          </div>
        )}

        {/* Sous-onglets */}
        <div className="flex items-center gap-2 pt-2 border-t border-slate-100">
          <button
            onClick={() => setActiveSubTab("chapters")}
            className={`px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2 cursor-pointer ${
              activeSubTab === "chapters"
                ? "bg-indigo-600 text-white shadow-xs"
                : "bg-slate-100 text-slate-600 hover:bg-slate-200"
            }`}
          >
            <FileText size={14} />
            <span>Chapitres du Mémoire ({sections.length})</span>
          </button>

          <button
            onClick={() => setActiveSubTab("personnel")}
            className={`px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2 cursor-pointer ${
              activeSubTab === "personnel"
                ? "bg-indigo-600 text-white shadow-xs"
                : "bg-slate-100 text-slate-600 hover:bg-slate-200"
            }`}
          >
            <Users size={14} />
            <span>Personnel Clé & CVs ({personnel.length})</span>
          </button>

          <button
            onClick={() => setActiveSubTab("materiel")}
            className={`px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2 cursor-pointer ${
              activeSubTab === "materiel"
                ? "bg-indigo-600 text-white shadow-xs"
                : "bg-slate-100 text-slate-600 hover:bg-slate-200"
            }`}
          >
            <Truck size={14} />
            <span>Matériel & Engins ({materiel.length})</span>
          </button>
        </div>
      </div>

      {/* VUE 1 : RÉDACTION DES CHAPITRES */}
      {activeSubTab === "chapters" && (
        <div className="grid lg:grid-cols-[320px_1fr] gap-6">
          {/* Sommaire latéral gauche */}
          <div className="bg-white border border-slate-200 rounded-2xl p-4 shadow-sm space-y-2 h-fit">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-500 px-2 mb-3">
              Sommaire du Mémoire
            </h3>

            <div className="space-y-1.5">
              {sections.map((sec, idx) => {
                const isSelected = sec.id === selectedSection?.id;
                return (
                  <button
                    key={sec.id}
                    onClick={() => handleSelectSection(sec)}
                    className={`w-full text-left p-3 rounded-xl text-xs transition-all flex items-start gap-2.5 cursor-pointer ${
                      isSelected
                        ? "bg-indigo-50 border border-indigo-300 text-indigo-950 font-bold shadow-xs"
                        : "hover:bg-slate-50 text-slate-700 border border-transparent"
                    }`}
                  >
                    <span
                      className={`w-5 h-5 rounded-full flex items-center justify-center text-[10px] shrink-0 font-mono ${
                        sec.statut === "VALIDE"
                          ? "bg-emerald-100 text-emerald-800"
                          : sec.statut === "GENERE"
                          ? "bg-blue-100 text-blue-800"
                          : "bg-slate-200 text-slate-600"
                      }`}
                    >
                      {idx + 1}
                    </span>
                    <div className="flex-1 min-w-0">
                      <p className="line-clamp-2 leading-snug">{sec.titre}</p>
                      <div className="mt-1 flex items-center gap-1.5">
                        <span
                          className={`inline-block px-1.5 py-0.2 rounded text-[9px] font-bold uppercase ${
                            sec.statut === "VALIDE"
                              ? "bg-emerald-100 text-emerald-700"
                              : sec.statut === "GENERE"
                              ? "bg-blue-100 text-blue-700"
                              : "bg-slate-100 text-slate-500"
                          }`}
                        >
                          {sec.statut === "VALIDE"
                            ? "Validé"
                            : sec.statut === "GENERE"
                            ? "Rédigé par IA"
                            : "À Rédiger"}
                        </span>
                      </div>
                    </div>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Éditeur / Lecteur de chapitre */}
          {selectedSection && (
            <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm space-y-5">
              {/* En-tête de section */}
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-4 border-b border-slate-100">
                <div>
                  <span className="text-[10px] font-extrabold uppercase tracking-wider text-indigo-600">
                    Chapitre en cours
                  </span>
                  <h3 className="text-base font-bold text-slate-900 mt-0.5">{selectedSection.titre}</h3>
                </div>

                <div className="flex items-center gap-2">
                  <button
                    onClick={() => {
                      setEditMode(!editMode);
                      if (!editMode) setCurrentContent(selectedSection.contenu);
                    }}
                    className={`px-3 py-1.5 rounded-xl text-xs font-semibold flex items-center gap-1.5 transition-colors cursor-pointer ${
                      editMode
                        ? "bg-slate-900 text-white"
                        : "bg-slate-100 hover:bg-slate-200 text-slate-700"
                    }`}
                  >
                    {editMode ? <Eye size={13} /> : <Edit3 size={13} />}
                    <span>{editMode ? "Aperçu" : "Modifier"}</span>
                  </button>

                  <button
                    onClick={handleValidateSection}
                    disabled={savingOffer || !selectedSection.contenu}
                    className="px-3 py-1.5 bg-emerald-50 hover:bg-emerald-100 text-emerald-700 border border-emerald-200 rounded-xl text-xs font-bold flex items-center gap-1 cursor-pointer transition-colors"
                  >
                    <Check size={13} />
                    <span>Valider</span>
                  </button>
                </div>
              </div>

              {/* Barre de consigne IA */}
              <div className="p-4 bg-slate-50 border border-slate-200/80 rounded-xl space-y-3">
                <div className="flex items-center gap-2">
                  <Sparkles size={16} className="text-indigo-600" />
                  <span className="text-xs font-bold text-slate-800">
                    Générateur IA spécialisé BTP & Marchés Publics
                  </span>
                </div>
                <div className="flex flex-col sm:flex-row items-center gap-2">
                  <input
                    type="text"
                    value={customPrompt}
                    onChange={(e) => setCustomPrompt(e.target.value)}
                    placeholder="Consigne spécifique (ex: Préciser l'épaisseur des pavés, les bordures T2 et les essais Proctor...)"
                    className="flex-1 w-full px-3.5 py-2 bg-white text-slate-900 placeholder:text-slate-400 border border-slate-300 rounded-xl text-xs focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  />
                  <button
                    onClick={handleGenerateSection}
                    disabled={generatingSection}
                    className="w-full sm:w-auto px-4 py-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white text-xs font-bold rounded-xl flex items-center justify-center gap-1.5 shadow-sm transition-all cursor-pointer shrink-0"
                  >
                    {generatingSection ? (
                      <>
                        <RefreshCw size={13} className="animate-spin" />
                        <span>Rédaction IA...</span>
                      </>
                    ) : (
                      <>
                        <Sparkles size={13} />
                        <span>{selectedSection.contenu ? "Régénérer" : "Rédiger ce chapitre"}</span>
                      </>
                    )}
                  </button>
                </div>
              </div>

              {saveMessage && (
                <div className="p-3 bg-blue-50 border border-blue-200 text-blue-900 text-xs rounded-xl flex items-center gap-2">
                  <CheckCircle2 size={15} className="text-blue-600 shrink-0" />
                  <span>{saveMessage}</span>
                </div>
              )}

              {/* Zone de contenu */}
              {editMode ? (
                <div className="space-y-3">
                  <textarea
                    rows={18}
                    value={currentContent}
                    onChange={(e) => setCurrentContent(e.target.value)}
                    className="w-full p-4 border border-slate-300 rounded-xl text-xs font-mono leading-relaxed focus:outline-none focus:ring-2 focus:ring-indigo-500 bg-white text-slate-900 placeholder:text-slate-400"
                  />
                  <div className="flex justify-end">
                    <button
                      onClick={handleSaveCurrentSectionText}
                      disabled={savingOffer}
                      className="px-4 py-2 bg-slate-900 hover:bg-slate-800 text-white text-xs font-bold rounded-xl flex items-center gap-1.5 shadow-sm cursor-pointer"
                    >
                      <Save size={13} />
                      <span>{savingOffer ? "Sauvegarde..." : "Enregistrer les modifications"}</span>
                    </button>
                  </div>
                </div>
              ) : selectedSection.contenu ? (
                <div className="w-full text-xs text-slate-900 leading-relaxed bg-white p-6 rounded-xl border border-slate-200 whitespace-pre-wrap font-sans select-text shadow-sm">
                  {selectedSection.contenu}
                </div>
              ) : (
                <div className="p-12 text-center text-xs text-slate-400 border-2 border-dashed border-slate-200 rounded-xl space-y-3">
                  <FileText size={32} className="mx-auto text-slate-300" />
                  <p className="font-semibold text-slate-600">Ce chapitre n'a pas encore été rédigé.</p>
                  <p className="text-slate-400 max-w-sm mx-auto">
                    Cliquez sur <b>« Rédiger ce chapitre »</b> ci-dessus pour laisser l'agent technique structurer une méthodologie détaillée.
                  </p>
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* VUE 2 : PERSONNEL CLÉ & CVS */}
      {activeSubTab === "personnel" && (
        <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm space-y-6">
          <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 pb-4 border-b border-slate-100">
            <div>
              <h3 className="text-sm font-bold text-slate-900 flex items-center gap-2">
                <Users size={16} className="text-indigo-600" />
                Affectation du Personnel Clé ({personnel.length})
              </h3>
              <p className="text-xs text-slate-500 mt-0.5">
                Associez vos experts d'entreprise aux profils exigés par le DAO ou laissez l'IA proposer et pré-remplir les profils.
              </p>
            </div>

            <div className="flex flex-wrap items-center gap-2">
              <button
                type="button"
                onClick={handleProposePersonnel}
                disabled={proposingPersonnel || personnel.length === 0}
                className="inline-flex items-center gap-1.5 px-4 py-2 bg-gradient-to-r from-indigo-600 to-blue-600 hover:from-indigo-700 hover:to-blue-700 disabled:opacity-50 text-white text-xs font-bold rounded-xl shadow-md shadow-indigo-500/20 transition-all cursor-pointer"
              >
                {proposingPersonnel ? (
                  <>
                    <RefreshCw size={13} className="animate-spin" />
                    <span>Recherche & Proposition IA...</span>
                  </>
                ) : (
                  <>
                    <Sparkles size={13} />
                    <span>Proposer & Remplir par l'IA</span>
                  </>
                )}
              </button>

              {personnel.length > 0 && (
                <>
                  <button
                    type="button"
                    onClick={validateAllPersonnel}
                    className="inline-flex items-center gap-1.5 px-3 py-2 bg-emerald-50 hover:bg-emerald-100 text-emerald-800 border border-emerald-200 text-xs font-bold rounded-xl transition-colors cursor-pointer"
                  >
                    <CheckCircle2 size={13} className="text-emerald-600" />
                    <span>Tout Valider</span>
                  </button>

                  <button
                    type="button"
                    onClick={handleCopyCVs}
                    disabled={copyingCVs || !personnel.some((p) => p.document_id)}
                    title="Copier automatiquement les fichiers PDF des CVs associés vers 03_Offre_Technique/Personnel_Cle_et_CV sur Google Drive"
                    className="inline-flex items-center gap-1.5 px-3 py-2 bg-indigo-50 hover:bg-indigo-100 disabled:opacity-50 text-indigo-800 border border-indigo-200 text-xs font-bold rounded-xl transition-colors cursor-pointer shadow-sm"
                  >
                    <FolderCheck size={13} className={copyingCVs ? "animate-spin text-indigo-600" : "text-indigo-600"} />
                    <span>{copyingCVs ? "Transfert en cours..." : "Transférer CVs vers Drive"}</span>
                  </button>
                </>
              )}

              <a
                href="/company"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 px-3 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-semibold rounded-xl transition-colors shrink-0"
              >
                <Building2 size={13} className="text-indigo-600" />
                <span>Coffre-fort RAG</span>
                <ExternalLink size={11} />
              </a>
            </div>
          </div>

          {cvCopySuccessMsg && (
            <div className="p-3 bg-emerald-50 border border-emerald-200 text-emerald-800 text-xs font-medium rounded-xl flex items-center justify-between gap-2 animate-in fade-in">
              <div className="flex items-center gap-2">
                <CheckCircle2 size={15} className="text-emerald-600 shrink-0" />
                <span>{cvCopySuccessMsg}</span>
              </div>
              <button
                type="button"
                onClick={() => setCvCopySuccessMsg(null)}
                className="text-xs text-emerald-600 hover:text-emerald-900 font-bold px-1.5 cursor-pointer"
              >
                ✕
              </button>
            </div>
          )}

          {personnel.length === 0 ? (
            <div className="p-8 text-center text-xs text-slate-400 border-2 border-dashed border-slate-200 rounded-xl">
              Aucun profil de personnel clé extrait automatiquement. Lancez l'analyse IA du DAO pour les détecter.
            </div>
          ) : (
            <div className="space-y-4">
              {personnel.map((p, idx) => (
                <div
                  key={idx}
                  className="p-4 border border-slate-200 rounded-xl bg-slate-50/50 space-y-3"
                >
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                    <span className="text-xs font-bold text-slate-900 flex items-center gap-2">
                      <span className="w-5 h-5 rounded-full bg-indigo-100 text-indigo-700 text-[10px] font-bold flex items-center justify-center">
                        {idx + 1}
                      </span>
                      {p.role_requis}
                    </span>

                    <div className="flex flex-wrap items-center gap-2">
                      {validatedPersonnel.includes(idx) ? (
                        <button
                          type="button"
                          onClick={() => toggleValidatePersonnel(idx)}
                          className="inline-flex items-center gap-1 text-[11px] font-bold text-emerald-700 bg-emerald-50 border border-emerald-200 px-2.5 py-0.5 rounded-full cursor-pointer hover:bg-emerald-100 transition-colors"
                        >
                          <CheckCircle2 size={12} className="text-emerald-600" />
                          <span>Profil Validé</span>
                        </button>
                      ) : p.nom ? (
                        <button
                          type="button"
                          onClick={() => toggleValidatePersonnel(idx)}
                          className="inline-flex items-center gap-1 text-[11px] font-bold text-amber-800 bg-amber-50 hover:bg-amber-100 border border-amber-200 px-2.5 py-0.5 rounded-full cursor-pointer transition-colors"
                        >
                          <Sparkles size={12} className="text-amber-600" />
                          <span>Proposition IA (Valider)</span>
                        </button>
                      ) : null}

                      {p.document_url && (
                        <a
                          href={p.document_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-[11px] font-semibold text-indigo-600 hover:underline flex items-center gap-1"
                        >
                          <span>Voir le CV sur Drive</span>
                          <ExternalLink size={10} />
                        </a>
                      )}
                    </div>
                  </div>

                  {/* Sélecteur de CV RAG */}
                  <div className="p-3 bg-indigo-50/70 border border-indigo-100 rounded-xl flex flex-col sm:flex-row sm:items-center justify-between gap-2.5">
                    <div className="flex items-center gap-2 text-xs font-semibold text-indigo-950">
                      <UserCheck size={15} className="text-indigo-600 shrink-0" />
                      <span>Lier un CV du Coffre-fort RAG :</span>
                    </div>
                    <div className="flex items-center gap-2 flex-1 sm:max-w-md">
                      <select
                        value={p.document_id || ""}
                        onChange={(e) => {
                          const docId = e.target.value;
                          const found = companyDocs.find((d) => d.google_drive_id === docId || d.id === docId);
                          if (found) {
                            handleUpdatePersonnel(idx, {
                              ...p,
                              nom: p.nom || found.titre,
                              document_id: found.google_drive_id || found.id,
                              document_url: found.web_view_link || "",
                              qualifications: p.qualifications || found.description || "",
                            });
                          } else if (docId === "") {
                            handleUpdatePersonnel(idx, {
                              ...p,
                              document_id: undefined,
                              document_url: undefined,
                            });
                          }
                        }}
                        className="w-full px-3 py-1.5 bg-white text-slate-900 border border-indigo-200 rounded-lg text-xs focus:outline-none focus:ring-2 focus:ring-indigo-500"
                      >
                        <option value="">-- Choisir un CV / profil existant --</option>
                        {companyDocs
                          .filter((d) => d.type_doc === "CV_EXPERT" || d.titre.toLowerCase().includes("cv") || d.type_doc === "AUTRE")
                          .map((doc) => (
                            <option key={doc.id} value={doc.google_drive_id || doc.id}>
                              {doc.titre} {doc.file_name ? `(${doc.file_name})` : ""}
                            </option>
                          ))}
                      </select>
                      {p.document_url && (
                        <a
                          href={p.document_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="px-2.5 py-1.5 bg-white hover:bg-indigo-100 text-indigo-700 border border-indigo-200 rounded-lg text-xs font-bold flex items-center gap-1 shrink-0 transition-colors"
                        >
                          <span>Voir CV</span>
                          <ExternalLink size={11} />
                        </a>
                      )}
                    </div>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                    <div>
                      <label className="block text-[10px] font-bold uppercase text-slate-500 mb-1">
                        Nom de l'expert
                      </label>
                      <input
                        type="text"
                        value={p.nom}
                        onChange={(e) => handleUpdatePersonnel(idx, { ...p, nom: e.target.value })}
                        placeholder="ex: Ing. Ibrahima Sarr"
                        className="w-full px-3 py-1.5 bg-white text-slate-900 placeholder:text-slate-400 border border-slate-300 rounded-lg text-xs"
                      />
                    </div>
                    <div>
                      <label className="block text-[10px] font-bold uppercase text-slate-500 mb-1">
                        Expérience (Années)
                      </label>
                      <input
                        type="number"
                        value={p.experience_annees || ""}
                        onChange={(e) =>
                          handleUpdatePersonnel(idx, {
                            ...p,
                            experience_annees: parseInt(e.target.value) || 0,
                          })
                        }
                        placeholder="ex: 10"
                        className="w-full px-3 py-1.5 bg-white text-slate-900 placeholder:text-slate-400 border border-slate-300 rounded-lg text-xs"
                      />
                    </div>
                    <div>
                      <label className="block text-[10px] font-bold uppercase text-slate-500 mb-1">
                        Diplômes & Qualifications
                      </label>
                      <input
                        type="text"
                        value={p.qualifications}
                        onChange={(e) =>
                          handleUpdatePersonnel(idx, { ...p, qualifications: e.target.value })
                        }
                        placeholder="ex: Master Génie Civil - Spécialiste Voirie"
                        className="w-full px-3 py-1.5 bg-white text-slate-900 placeholder:text-slate-400 border border-slate-300 rounded-lg text-xs"
                      />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* VUE 3 : MATÉRIEL & ENGINS */}
      {activeSubTab === "materiel" && (
        <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm space-y-6">
          <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 pb-4 border-b border-slate-100">
            <div>
              <h3 className="text-sm font-bold text-slate-900 flex items-center gap-2">
                <Truck size={16} className="text-indigo-600" />
                Matériel & Engins Mobilisés ({materiel.length})
              </h3>
              <p className="text-xs text-slate-500 mt-0.5">
                Renseignez les engins alloués au chantier ou laissez l'IA proposer les modèles optimaux.
              </p>
            </div>

            <div className="flex flex-wrap items-center gap-2">
              <button
                type="button"
                onClick={handleProposeMateriel}
                disabled={proposingMateriel || materiel.length === 0}
                className="inline-flex items-center gap-1.5 px-4 py-2 bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-700 hover:to-indigo-700 disabled:opacity-50 text-white text-xs font-bold rounded-xl shadow-md shadow-purple-500/20 transition-all cursor-pointer"
              >
                {proposingMateriel ? (
                  <>
                    <RefreshCw size={13} className="animate-spin" />
                    <span>Allocation IA...</span>
                  </>
                ) : (
                  <>
                    <Sparkles size={13} />
                    <span>Proposer les engins par l'IA</span>
                  </>
                )}
              </button>

              {materiel.length > 0 && (
                <button
                  type="button"
                  onClick={validateAllMateriel}
                  className="inline-flex items-center gap-1.5 px-3 py-2 bg-emerald-50 hover:bg-emerald-100 text-emerald-800 border border-emerald-200 text-xs font-bold rounded-xl transition-colors cursor-pointer"
                >
                  <CheckCircle2 size={13} className="text-emerald-600" />
                  <span>Tout Valider</span>
                </button>
              )}

              <a
                href="/company"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 px-3 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-semibold rounded-xl transition-colors shrink-0"
              >
                <Building2 size={13} className="text-indigo-600" />
                <span>Coffre-fort RAG</span>
                <ExternalLink size={11} />
              </a>
            </div>
          </div>

          {materiel.length === 0 ? (
            <div className="p-8 text-center text-xs text-slate-400 border-2 border-dashed border-slate-200 rounded-xl">
              Aucun matériel extrait automatiquement. Lancez l'analyse IA du DAO pour les détecter.
            </div>
          ) : (
            <div className="space-y-4">
              {materiel.map((m, idx) => (
                <div
                  key={idx}
                  className="p-4 border border-slate-200 rounded-xl bg-slate-50/50 space-y-3"
                >
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                    <div className="flex items-center gap-2">
                      <span className="w-5 h-5 rounded-full bg-indigo-100 text-indigo-700 text-[10px] font-bold flex items-center justify-center">
                        {idx + 1}
                      </span>
                      <span className="text-xs font-bold text-slate-900">{m.type_requis}</span>
                    </div>

                    <div className="flex items-center gap-2">
                      {validatedMateriel.includes(idx) ? (
                        <button
                          type="button"
                          onClick={() => toggleValidateMateriel(idx)}
                          className="inline-flex items-center gap-1 text-[11px] font-bold text-emerald-700 bg-emerald-50 border border-emerald-200 px-2.5 py-0.5 rounded-full cursor-pointer hover:bg-emerald-100 transition-colors"
                        >
                          <CheckCircle2 size={12} className="text-emerald-600" />
                          <span>Matériel Validé</span>
                        </button>
                      ) : m.designation ? (
                        <button
                          type="button"
                          onClick={() => toggleValidateMateriel(idx)}
                          className="inline-flex items-center gap-1 text-[11px] font-bold text-purple-800 bg-purple-50 hover:bg-purple-100 border border-purple-200 px-2.5 py-0.5 rounded-full cursor-pointer transition-colors"
                        >
                          <Sparkles size={12} className="text-purple-600" />
                          <span>Proposition IA (Valider)</span>
                        </button>
                      ) : null}
                    </div>
                  </div>

                  {/* Sélecteur de matériel RAG */}
                  <div className="p-3 bg-purple-50/70 border border-purple-100 rounded-xl flex flex-col sm:flex-row sm:items-center justify-between gap-2.5">
                    <div className="flex items-center gap-2 text-xs font-semibold text-purple-950">
                      <Truck size={15} className="text-purple-600 shrink-0" />
                      <span>Lier un engin / équipement du Coffre-fort RAG :</span>
                    </div>
                    <div className="flex items-center gap-2 flex-1 sm:max-w-md">
                      <select
                        onChange={(e) => {
                          const docId = e.target.value;
                          const found = companyDocs.find((d) => d.google_drive_id === docId || d.id === docId);
                          if (found) {
                            handleUpdateMateriel(idx, {
                              ...m,
                              designation: found.titre,
                            });
                          }
                        }}
                        className="w-full px-3 py-1.5 bg-white text-slate-900 border border-purple-200 rounded-lg text-xs focus:outline-none focus:ring-2 focus:ring-purple-500"
                      >
                        <option value="">-- Choisir un équipement du Coffre-fort --</option>
                        {companyDocs
                          .filter((d) => d.type_doc === "JUSTIFICATIF_MATERIEL" || d.type_doc === "AUTRE")
                          .map((doc) => (
                            <option key={doc.id} value={doc.google_drive_id || doc.id}>
                              {doc.titre} {doc.file_name ? `(${doc.file_name})` : ""}
                            </option>
                          ))}
                      </select>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                    <div>
                      <label className="block text-[10px] font-bold uppercase text-slate-500 mb-1">
                        Désignation / Marque & Modèle
                      </label>
                      <input
                        type="text"
                        value={m.designation}
                        onChange={(e) => handleUpdateMateriel(idx, { ...m, designation: e.target.value })}
                        placeholder="ex: Compacteur BOMAG BW 120"
                        className="w-full px-3 py-1.5 bg-white text-slate-900 placeholder:text-slate-400 border border-slate-300 rounded-lg text-xs"
                      />
                    </div>
                    <div>
                      <label className="block text-[10px] font-bold uppercase text-slate-500 mb-1">
                        Immatriculation / N° Série
                      </label>
                      <input
                        type="text"
                        value={m.immatriculation || ""}
                        onChange={(e) =>
                          handleUpdateMateriel(idx, { ...m, immatriculation: e.target.value })
                        }
                        placeholder="ex: SN-DK-2024-X"
                        className="w-full px-3 py-1.5 bg-white text-slate-900 placeholder:text-slate-400 border border-slate-300 rounded-lg text-xs"
                      />
                    </div>
                    <div>
                      <label className="block text-[10px] font-bold uppercase text-slate-500 mb-1">
                        Statut Juridique
                      </label>
                      <select
                        value={m.statut}
                        onChange={(e) =>
                          handleUpdateMateriel(idx, {
                            ...m,
                            statut: e.target.value as "PROPRIETE" | "LOCATION" | "LEASING",
                          })
                        }
                        className="w-full px-3 py-1.5 bg-white text-slate-900 border border-slate-300 rounded-lg text-xs"
                      >
                        <option value="PROPRIETE">Propriété de l'entreprise</option>
                        <option value="LOCATION">En location (Attestation/Contrat)</option>
                        <option value="LEASING">Leasing / Crédit-Bail</option>
                      </select>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
