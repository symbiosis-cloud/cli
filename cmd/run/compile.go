package run

import (
	"log"
)

type BuildEngine string

const (
	ENGINE_HELM      BuildEngine = "helm"
	ENGINE_KUSTOMIZE BuildEngine = "kustomize"
)

func compileManifests(runFile *RunFile, buildEngine BuildEngine) {
	log.Printf("Compiling manifests using build engine: %s", buildEngine)

}
