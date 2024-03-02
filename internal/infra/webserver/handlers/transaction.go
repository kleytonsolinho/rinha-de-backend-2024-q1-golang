package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kleytonsolinho/rinha-de-backend-2024-q1/internal/infra/database"
	"github.com/kleytonsolinho/rinha-de-backend-2024-q1/internal/infra/dto"
)

func TransactionHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	userId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || userId < 1 || userId > 5 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("User ID is invalid")
		return
	}

	var transaction dto.TransactionInputDTO
	err = json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validTransaction, err := TransactionValidator(transaction.Valor, transaction.Tipo, transaction.Descricao)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := NewTransaction(
		validTransaction.Valor,
		validTransaction.Tipo,
		validTransaction.Descricao,
		userId,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func TransactionValidator(value int64, tipo string, description string) (*dto.TransactionInputDTO, error) {
	if value <= 0 {
		return nil, errors.New("valor da transação precisa ser positivo")
	}
	if tipo != "c" && tipo != "d" {
		return nil, errors.New("tipo da transação inválido")
	}
	if len(description) < 1 || len(description) > 10 {
		return nil, errors.New("descrição da transação deve ter entre 1 e 10 caracteres")
	}

	return &dto.TransactionInputDTO{
		Valor:     value,
		Tipo:      tipo,
		Descricao: description,
	}, nil
}

func NewTransaction(valor int64, tipo string, descricao string, userId int64) (*dto.TransactionOutputDTO, error) {
	db, err := database.NewPostgresStorage()
	if err != nil {
		return nil, err
	}

	valueInCents := valor * 100

	fmt.Println(&dto.TransactionInputDTO{
		Valor:     valueInCents,
		Tipo:      tipo,
		Descricao: descricao,
		ClienteID: userId,
	})

	var transactionCreated *dto.TransactionDTO
	transactionCreated, err = db.CreateTransaction(&dto.TransactionInputDTO{
		Valor:     valueInCents,
		Tipo:      tipo,
		Descricao: descricao,
		ClienteID: userId,
	})
	if err != nil {
		return nil, err
	}

	fmt.Println(transactionCreated)

	return &dto.TransactionOutputDTO{
		Limite: 10000,
		Saldo:  10000 - valueInCents,
	}, nil
}
